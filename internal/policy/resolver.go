package policy

import (
	"time"

	common "github.com/itsthatdude/jit-access-controller/internal/common"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

type PolicyResolver struct{}

func (r *PolicyResolver) Resolve(
	req common.AccessRequestObject,
	policies []common.AccessPolicyObject,
) common.AccessPolicyObject {

	for i := range policies {
		if matchesPolicy(policies[i], req) {
			return policies[i]
		}
	}

	return nil
}

func matchesPolicy(
	policy common.AccessPolicyObject,
	req common.AccessRequestObject,
) bool {
	var policySpec = policy.GetPolicy()
	var reqSpec = req.GetSpec()
	return matchesSubjects(policySpec.Requesters, reqSpec.Subject, reqSpec.Groups) &&
		matchesDuration(policySpec.MaxDuration, reqSpec.Duration) &&
		matchesPermissions(policySpec.AllowedPermissions, reqSpec.Permissions) &&
		matchesRoles(policySpec.AllowedRoles, reqSpec.Role)
}

func matchesSubjects(
	allowedSubjects []rbacv1.Subject,
	subject string,
	groups []string,
) bool {
	if len(allowedSubjects) == 0 {
		return true
	}

	for _, sub := range allowedSubjects {
		if sub.Kind == rbacv1.UserKind && sub.Name == subject {
			return true
		}

		if sub.Kind == rbacv1.GroupKind {
			for _, group := range groups {
				if group == sub.Name {
					return true
				}
			}
		}
	}

	return false
}

func matchesDuration(
	policyMaxDuration string,
	requestDuration string,
) bool {
	specDuration, err := time.ParseDuration(requestDuration)
	if err != nil {
		return false
	}

	maxDuration, err := time.ParseDuration(policyMaxDuration)
	if err != nil {
		return false
	}

	if specDuration < maxDuration {
		return true
	}

	return false
}

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

func policyRuleIsSubsetWithWildcard(requested, allowed rbacv1.PolicyRule) bool {
	return fieldAllows(requested.APIGroups, allowed.APIGroups) &&
		fieldAllows(requested.Resources, allowed.Resources) &&
		fieldAllows(requested.ResourceNames, allowed.ResourceNames) &&
		fieldAllows(requested.Verbs, allowed.Verbs) &&
		fieldAllows(requested.NonResourceURLs, allowed.NonResourceURLs)
}

func matchesPermissions(
	allowedRules,
	requestedRules []rbacv1.PolicyRule,
) bool {
	if len(requestedRules) == 0 {
		return true
	}

	for _, reqRule := range requestedRules {
		matched := false
		for _, allowRule := range allowedRules {
			if policyRuleIsSubsetWithWildcard(reqRule, allowRule) {
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

func matchesRoles(allowedRoles []rbacv1.RoleRef, requestedRole rbacv1.RoleRef) bool {
	for _, ref := range allowedRoles {
		if roleRefEquals(ref, requestedRole) {
			return true
		}
	}
	return false
}
