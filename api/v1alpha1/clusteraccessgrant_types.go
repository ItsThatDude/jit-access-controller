package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *ClusterAccessGrant) GetStatus() *AccessGrantStatus {
	return &r.Status
}
func (r *ClusterAccessGrant) SetStatus(status *AccessGrantStatus) {
	r.Status = *status
}
func (r *ClusterAccessGrant) GetScope() RequestScope {
	return RequestScopeCluster
}
func (r *ClusterAccessGrant) GetName() string {
	return r.Name
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// ClusterAccessGrant is the Schema for the clusteraccessgrants API
// +kubebuilder:printcolumn:name="Subject",type=string,JSONPath=`.status.subject`
// +kubebuilder:printcolumn:name="Access-Expires-At",type=string,JSONPath=`.status.accessExpiresAt`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type ClusterAccessGrant struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// status defines the observed state of ClusterAccessGrant
	// +optional
	Status AccessGrantStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// ClusterAccessGrantList contains a list of ClusterAccessGrant
type ClusterAccessGrantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []ClusterAccessGrant `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &ClusterAccessGrant{}, &ClusterAccessGrantList{})
		return nil
	})
}
