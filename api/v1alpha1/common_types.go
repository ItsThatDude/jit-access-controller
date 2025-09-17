package v1alpha1

type RequestState string

const (
	RequestStatePending  RequestState = "Pending"
	RequestStateApproved RequestState = "Approved"
	RequestStateDenied   RequestState = "Denied"
)

type ResponseState string

const (
	ResponseStateApproved ResponseState = "Approved"
	ResponseStateDenied   ResponseState = "Denied"
)

type RoleKind string

const (
	RoleKindRole        RoleKind = "Role"
	RoleKindClusterRole RoleKind = "ClusterRole"
)
