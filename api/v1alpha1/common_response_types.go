package v1alpha1

// +kubebuilder:validation:Enum=Approved;Denied
type ResponseState string

const (
	ResponseStateApproved ResponseState = "Approved"
	ResponseStateDenied   ResponseState = "Denied"
)

type JITAccessResponseSpec struct {
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="RequestRef cannot be changed after creation"
	RequestRef string `json:"requestRef"`

	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Approver cannot be changed after creation"
	Approver string `json:"approver"`

	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Response cannot be changed after creation"
	Response ResponseState `json:"response"`
}

type JITAccessResponseStatus struct {
}
