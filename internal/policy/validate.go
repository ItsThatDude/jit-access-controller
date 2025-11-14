package policy

import (
	"context"
	"slices"
	"time"

	accessv1alpha1 "github.com/itsthatdude/jit-access-controller/api/v1alpha1"
	common "github.com/itsthatdude/jit-access-controller/internal/common"
	"github.com/itsthatdude/jit-access-controller/internal/utils"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func normalizeAPIGroup(s string) string {
	if s == "" {
		return "rbac.authorization.k8s.io"
	}
	return s
}

func roleRefEquals(a, b rbacv1.RoleRef) bool {
	return normalizeAPIGroup(a.APIGroup) == normalizeAPIGroup(b.APIGroup) &&
		a.Kind == b.Kind &&
		a.Name == b.Name
}

func roleRefSliceContains(slice []rbacv1.RoleRef, r rbacv1.RoleRef) bool {
	for _, ref := range slice {
		if roleRefEquals(ref, r) {
			return true
		}
	}
	return false
}

func fieldAllows(requested, allowed []string) bool {
	if hasWildcard(allowed) {
		return true
	}
	return sets.NewString(allowed...).HasAll(requested...)
}

func hasWildcard(slice []string) bool {
	for _, s := range slice {
		if s == "*" {
			return true
		}
	}
	return false
}

func RuleIsSubsetWithWildcard(requested, allowed rbacv1.PolicyRule) bool {
	return fieldAllows(requested.APIGroups, allowed.APIGroups) &&
		fieldAllows(requested.Resources, allowed.Resources) &&
		fieldAllows(requested.ResourceNames, allowed.ResourceNames) &&
		fieldAllows(requested.Verbs, allowed.Verbs) &&
		fieldAllows(requested.NonResourceURLs, allowed.NonResourceURLs)
}

func AllRequestedPolicyRulesAllowed(requestedRules, allowedRules []rbacv1.PolicyRule) bool {
	for _, reqRule := range requestedRules {
		matched := false
		for _, allowRule := range allowedRules {
			if RuleIsSubsetWithWildcard(reqRule, allowRule) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func IsRequestValid[T common.AccessPolicyObject](
	req common.AccessRequestObject,
	policies []T,
) (bool, *accessv1alpha1.SubjectPolicy) {
	log := logf.FromContext(context.Background())
	for _, p := range policies {
		policy := p.GetPolicy()
		spec := req.GetSpec()

		userNames := []string{}
		groupNames := []string{}
		for _, approver := range policy.Requesters {
			switch approver.Kind {
			case rbacv1.UserKind:
				userNames = append(userNames, approver.Name)
			case rbacv1.GroupKind:
				groupNames = append(groupNames, approver.Name)
			}
		}

		group_matched := utils.SliceOverlaps(groupNames, spec.Groups)
		user_matched := slices.Contains(userNames, spec.Subject)

		// Subject or group must match
		if !group_matched && !user_matched {
			continue
		}

		// Duration must be within policy threshold
		specDuration, err := time.ParseDuration(spec.Duration)
		if err != nil {
			log.Error(err, "failed to parse spec duration")
			continue
		}

		maxDuration, err := time.ParseDuration(policy.MaxDuration)
		if err != nil {
			log.Error(err, "failed to parse policy MaxDuration")
			continue
		}

		if specDuration > maxDuration {
			continue
		}

		// Permissions check (empty requested permissions are always allowed)
		permissionsAllowed := len(spec.Permissions) == 0 ||
			AllRequestedPolicyRulesAllowed(spec.Permissions, policy.AllowedPermissions)

		// Role check (empty role means "no role requested", so skip check)
		roleAllowed := spec.Role.Name == "" ||
			roleRefSliceContains(policy.AllowedRoles, spec.Role)

		if permissionsAllowed && roleAllowed {
			return true, &policy
		}
	}
	return false, nil
}
