package policy

import (
	accessv1alpha1 "antware.xyz/jitaccess/api/v1alpha1"
	"antware.xyz/jitaccess/internal/utils"
)

func ValidateNamespaced(jit *accessv1alpha1.JITAccessRequest, policies *accessv1alpha1.JITAccessPolicyList) bool {
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

func ValidateCluster(jit *accessv1alpha1.ClusterJITAccessRequest, policies *accessv1alpha1.ClusterJITAccessPolicyList) bool {
	permitted := false

	for _, policy := range policies.Items {
		for _, p := range policy.Spec.Policies {
			if p.Subject == jit.Spec.Subject {
				if utils.Contains(p.AllowedRoles, jit.Spec.Role) &&
					utils.Contains(p.AllowedNamespaces, jit.Namespace) &&
					jit.Spec.DurationSeconds <= p.MaxDurationSeconds {
					permitted = true
					break
				}
			}
		}
	}

	return permitted
}
