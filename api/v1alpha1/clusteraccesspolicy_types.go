package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r ClusterAccessPolicy) GetName() string {
	return r.Name
}

func (r ClusterAccessPolicy) GetNamespace() string {
	return ""
}

func (r ClusterAccessPolicy) GetScope() PolicyScope {
	return "Cluster"
}

func (r ClusterAccessPolicy) GetPolicy() SubjectPolicy {
	return r.Spec.SubjectPolicy
}

// ClusterAccessPolicySpec defines the desired state of ClusterAccessPolicy
type ClusterAccessPolicySpec struct {
	SubjectPolicy `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// ClusterAccessPolicy is the Schema for the clusteraccesspolicies API
type ClusterAccessPolicy struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of ClusterAccessPolicy
	// +required
	Spec ClusterAccessPolicySpec `json:"spec"`

	// status defines the observed state of ClusterAccessPolicy
	// +optional
	Status AccessPolicyStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// ClusterAccessPolicyList contains a list of ClusterAccessPolicy
type ClusterAccessPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []ClusterAccessPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &ClusterAccessPolicy{}, &ClusterAccessPolicyList{})
		return nil
	})
}
