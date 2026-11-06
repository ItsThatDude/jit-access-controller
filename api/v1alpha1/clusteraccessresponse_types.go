package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *ClusterAccessResponse) GetResponse() ResponseState {
	return r.Spec.Response
}

func (r *ClusterAccessResponse) GetApprover() string {
	return r.Spec.Approver
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// ClusterAccessResponse is the Schema for the clusteraccessresponses API
type ClusterAccessResponse struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of ClusterAccessResponse
	// +required
	Spec AccessResponseSpec `json:"spec"`

	// status defines the observed state of ClusterAccessResponse
	// +optional
	Status AccessResponseStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// ClusterAccessResponseList contains a list of ClusterAccessResponse
type ClusterAccessResponseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []ClusterAccessResponse `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &ClusterAccessResponse{}, &ClusterAccessResponseList{})
		return nil
	})
}
