package v1alpha1

import (
	rbacv1 "k8s.io/api/rbac/v1"
)

// SubjectPolicy defines access rules for a single subject (user/serviceaccount).
type SubjectPolicy struct {
	// The permitted users and groups that can request resources under this policy.
	// +required
	// +kubebuilder:validation:MinItems=1
	Requesters []rbacv1.Subject `json:"requesters,omitempty"`

	// AllowedRoles is a list of roles the subject is allowed to request.
	AllowedRoles []rbacv1.RoleRef `json:"allowedRoles,omitempty"`

	// AllowedPermissions is a list of adhoc permissions the subject is allowed to request.
	AllowedPermissions []rbacv1.PolicyRule `json:"allowedPermissions,omitempty"`

	// Duration specifies the maximum amount of time the access can last (e.g. "5s", "10m", "2h45m").
	// +kubebuilder:validation:Pattern=`^(\d+(ns|us|µs|ms|s|m|h))+$`
	// +required
	MaxDuration string `json:"maxDuration"`

	// The minimum number of approvals required to grant the request
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default:=1
	RequiredApprovals int `json:"requiredApprovals"`

	// The users and groups allowed to approve requests for this subject
	// +required
	// +kubebuilder:validation:MinItems=1
	Approvers []rbacv1.Subject `json:"approvers,omitempty"`
}
