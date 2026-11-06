package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *AccessGrant) GetStatus() *AccessGrantStatus {
	return &r.Status
}
func (r *AccessGrant) SetStatus(status *AccessGrantStatus) {
	r.Status = *status
}
func (r *AccessGrant) GetScope() RequestScope {
	return RequestScopeNamespace
}
func (r *AccessGrant) GetName() string {
	return r.Name
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// AccessGrant is the Schema for the accessgrants API
// +kubebuilder:printcolumn:name="Subject",type=string,JSONPath=`.status.subject`
// +kubebuilder:printcolumn:name="Access-Expires-At",type=string,JSONPath=`.status.accessExpiresAt`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type AccessGrant struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// status defines the observed state of AccessGrant
	// +optional
	Status AccessGrantStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// AccessGrantList contains a list of AccessGrant
type AccessGrantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []AccessGrant `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &AccessGrant{}, &AccessGrantList{})
		return nil
	})
}
