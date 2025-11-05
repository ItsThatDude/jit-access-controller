package v1alpha1

// +kubebuilder:validation:Enum=Approved;Denied
type ResponseState string

const (
	ResponseStateApproved ResponseState = "Approved"
	ResponseStateDenied   ResponseState = "Denied"
)

type AccessResponseSpec struct {
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="RequestRef cannot be changed after creation"
	// +required
	RequestRef string `json:"requestRef"`

	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Approver cannot be changed after creation"
	// +required
	Approver string `json:"approver"`

	// Groups are the groups the approver belongs to
	// +optional
	// +listType=set
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Groups cannot be changed after creation"
	Groups []string `json:"groups,omitempty"`

	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Response cannot be changed after creation"
	// +required
	Response ResponseState `json:"response"`
}

type AccessResponseStatus struct {
}
