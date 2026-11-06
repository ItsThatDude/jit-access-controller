package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *AccessResponse) GetResponse() ResponseState {
	return r.Spec.Response
}

func (r *AccessResponse) GetApprover() string {
	return r.Spec.Approver
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// AccessResponse is the Schema for the accessresponses API
type AccessResponse struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of AccessResponse
	// +required
	Spec AccessResponseSpec `json:"spec"`

	// status defines the observed state of AccessResponse
	// +optional
	Status AccessResponseStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// AccessResponseList contains a list of AccessResponse
type AccessResponseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []AccessResponse `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &AccessResponse{}, &AccessResponseList{})
		return nil
	})
}
