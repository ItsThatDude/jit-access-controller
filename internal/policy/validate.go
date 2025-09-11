package policy

import (
	accessv1alpha1 "antware.xyz/jitaccess/api/v1alpha1"
	"antware.xyz/jitaccess/internal/utils"
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

func IsNamespacedRequestValid(
	jit *accessv1alpha1.JITAccessRequest,
	policies *accessv1alpha1.JITAccessPolicyList,
) bool {
	for _, item := range policies.Items {
		for _, policy := range item.Spec.Policies {
			// Subject must match
			if !utils.Contains(policy.Subjects, jit.Spec.Subject) {
				continue
			}

			// Permissions check (empty requested permissions are always allowed)
			permissionsAllowed := len(jit.Spec.Permissions) == 0 ||
				AllRequestedPolicyRulesAllowed(jit.Spec.Permissions, policy.AllowedPermissions)

			// Role check (empty role means "no role requested", so skip check)
			roleAllowed := jit.Spec.Role == "" ||
				(utils.Contains(policy.AllowedRoles, jit.Spec.Role) &&
					jit.Spec.DurationSeconds <= policy.MaxDurationSeconds)

			if permissionsAllowed && roleAllowed {
				return true
			}
		}
	}
	return false
}

func IsClusterRequestValid(
	jit *accessv1alpha1.ClusterJITAccessRequest,
	policies *accessv1alpha1.ClusterJITAccessPolicyList,
) bool {
	for _, item := range policies.Items {
		for _, policy := range item.Spec.Policies {
			// Subject must match
			if !utils.Contains(policy.Subjects, jit.Spec.Subject) {
				continue
			}

			// Permissions check (empty requested permissions are always allowed)
			permissionsAllowed := len(jit.Spec.Permissions) == 0 ||
				AllRequestedPolicyRulesAllowed(jit.Spec.Permissions, policy.AllowedPermissions)

			// Role check (empty role means "no role requested", so skip check)
			roleAllowed := jit.Spec.ClusterRole == "" ||
				(utils.Contains(policy.AllowedClusterRoles, jit.Spec.ClusterRole) &&
					jit.Spec.DurationSeconds <= policy.MaxDurationSeconds)

			if permissionsAllowed && roleAllowed {
				return true
			}
		}
	}
	return false
}
