package policy

import (
	"slices"
	"time"

	accessv1alpha1 "antware.xyz/jitaccess/api/v1alpha1"
	common "antware.xyz/jitaccess/internal/common"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

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

func ruleIsSubsetWithWildcard(requested, allowed rbacv1.PolicyRule) bool {
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
			if ruleIsSubsetWithWildcard(reqRule, allowRule) {
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

func IsRequestValid[T common.AccessPolicyListInterface](
	req common.AccessRequestObject,
	policies []T,
) (bool, *accessv1alpha1.SubjectPolicy) {
	for _, p := range policies {
		policy := p.GetPolicy()
		spec := req.GetSpec()

		// Subject must match
		if !slices.Contains(policy.Subjects, spec.Subject) {
			continue
		}

		// Duration must be within policy threshold
		specDuration, err := time.ParseDuration(spec.Duration)
		if err != nil {
			continue
		}

		maxDuration, err := time.ParseDuration(policy.MaxDuration)
		if err != nil {
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
			slices.Contains(policy.AllowedRoles, spec.Role)

		if permissionsAllowed && roleAllowed {
			return true, &policy
		}
	}
	return false, nil
}
