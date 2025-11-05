package v1alpha1

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Enum=Cluster;Namespace
type RequestScope string

const (
	RequestScopeCluster   RequestScope = "Cluster"
	RequestScopeNamespace RequestScope = "Namespace"
)

// +kubebuilder:validation:Enum=Pending;Approved;Denied;Expired
type RequestState string

const (
	RequestStatePending  RequestState = "Pending"
	RequestStateApproved RequestState = "Approved"
	RequestStateDenied   RequestState = "Denied"
	RequestStateExpired  RequestState = "Expired"
)

type AccessRequestBaseSpec struct {
	// Subject is the username or identity requesting access
	// +required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Subject cannot be changed after creation"
	Subject string `json:"subject"`

	// Groups are the groups the subject belongs to
	// +optional
	// +listType=set
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Groups cannot be changed after creation"
	Groups []string `json:"groups,omitempty"`

	// Role is an optional pre-defined Role/ClusterRole to bind
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Role cannot be changed after creation"
	Role rbacv1.RoleRef `json:"role,omitempty"`

	// Permissions are adhoc RBAC rules to grant (instead of a pre-defined role)
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Permissions cannot be changed after creation"
	Permissions []rbacv1.PolicyRule `json:"permissions,omitempty"`

	// Duration specifies how long the access should last (e.g. "5s", "10m", "2h45m").
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Duration cannot be changed after creation"
	// +kubebuilder:validation:Pattern=`^(\d+(ns|us|µs|ms|s|m|h))+$`
	// +kubebuilder:default:="10m"
	Duration string `json:"duration"`

	// User's justification for the request
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Justification cannot be changed after creation"
	Justification string `json:"justification"`
}

type AccessRequestStatus struct {
	RequestId         string       `json:"requestId,omitempty"`
	State             RequestState `json:"state,omitempty"`
	ApprovalsRequired int          `json:"approvalsRequired,omitempty"`
	ApprovalsReceived int          `json:"approvalsReceived,omitempty"`
	RequestExpiresAt  *metav1.Time `json:"requestExpiresAt,omitempty"`

	GrantCreated bool `json:"grantCreated,omitempty"`
}
