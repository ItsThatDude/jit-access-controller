package v1alpha1

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Enum=Pending;Approved;Denied
type RequestState string

const (
	RequestStatePending  RequestState = "Pending"
	RequestStateApproved RequestState = "Approved"
	RequestStateDenied   RequestState = "Denied"
)

// +kubebuilder:validation:Enum=Role;ClusterRole
type RoleKind string

const (
	RoleKindRole        RoleKind = "Role"
	RoleKindClusterRole RoleKind = "ClusterRole"
)

type JITAccessRequestBaseSpec struct {
	// Subject is the username or identity requesting access
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Subject cannot be changed after creation"
	Subject string `json:"subject"`

	// Role is an optional pre-defined Role/ClusterRole to bind
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Role cannot be changed after creation"
	Role string `json:"role,omitempty"`

	// Permissions are adhoc RBAC rules to grant (instead of a pre-defined role)
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Permissions cannot be changed after creation"
	Permissions []rbacv1.PolicyRule `json:"permissions,omitempty"`

	// DurationSeconds defines how long the access should last
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="DurationSeconds cannot be changed after creation"
	DurationSeconds int64 `json:"durationSeconds"`

	// User's justification for the request
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Justification cannot be changed after creation"
	Justification string `json:"justification"`
}

type JITAccessRequestStatus struct {
	RequestId         string       `json:"requestId,omitempty"`
	State             RequestState `json:"state,omitempty"`
	ApprovalsRequired int          `json:"approvalsRequired,omitempty"`
	ApprovalsReceived int          `json:"approvalsReceived,omitempty"`
	ExpiresAt         *metav1.Time `json:"expiresAt,omitempty"`

	RoleBindingCreated      bool `json:"roleBindingCreated,omitempty"`
	AdhocRoleCreated        bool `json:"adhocRoleCreated,omitempty"`
	AdhocRoleBindingCreated bool `json:"adhocRoleBindingCreated,omitempty"`
}
