package v1alpha1

import (
	rbacv1 "k8s.io/api/rbac/v1"
)

// SubjectPolicy defines access rules for a single subject (user/serviceaccount).
type SubjectPolicy struct {
	// Subject is the identity (email or K8s username) of the requester.
	Subjects []string `json:"subjects"`

	// AllowedRoles is a list of roles the subject is allowed to request.
	AllowedRoles []rbacv1.RoleRef `json:"allowedRoles,omitempty"`

	// AllowedPermissions is a list of adhoc permissions the subject is allowed to request.
	AllowedPermissions []rbacv1.PolicyRule `json:"allowedPermissions,omitempty"`

	// Duration specifies the maximum amount of time the access can last (e.g. "5s", "10m", "2h45m").
	// +kubebuilder:validation:Pattern=`^(\d+(ns|us|µs|ms|s|m|h))+$`
	MaxDuration string `json:"maxDuration"`

	// The minimum number of approvals required to grant the request
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default:=1
	RequiredApprovals int `json:"requiredApprovals"`

	// Approvers
	Approvers []string `json:"approvers,omitempty"`

	// Approver Groups
	ApproverGroups []string `json:"approverGroups,omitempty"`
}
