package policy

import (
	accessv1alpha1 "antware.xyz/jitaccess/api/v1alpha1"
	"antware.xyz/jitaccess/internal/utils"
)

func IsNamespacedRequestValid(jit *accessv1alpha1.JITAccessRequest, policies *accessv1alpha1.JITAccessPolicyList) bool {
	permitted := false

	for _, policy := range policies.Items {
		for _, p := range policy.Spec.Policies {
			if p.Subject == jit.Spec.Subject {
				if utils.Contains(p.AllowedRoles, jit.Spec.Role) &&
					jit.Spec.DurationSeconds <= p.MaxDurationSeconds {
					permitted = true
					break
				}
			}
		}
	}

	return permitted
}

func IsClusterRequestValid(jit *accessv1alpha1.ClusterJITAccessRequest, policies *accessv1alpha1.ClusterJITAccessPolicyList) bool {
	permitted := false

	for _, policy := range policies.Items {
		for _, p := range policy.Spec.Policies {
			if p.Subject == jit.Spec.Subject {
				if utils.Contains(p.AllowedClusterRoles, jit.Spec.ClusterRole) &&
					jit.Spec.DurationSeconds <= p.MaxDurationSeconds {
					permitted = true
					break
				}
			}
		}
	}

	return permitted
}
