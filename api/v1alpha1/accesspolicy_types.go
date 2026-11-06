package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r AccessPolicy) GetName() string {
	return r.Name
}

func (r AccessPolicy) GetNamespace() string {
	return r.Namespace
}

func (r AccessPolicy) GetScope() PolicyScope {
	return "Namespace"
}

func (r AccessPolicy) GetPolicy() SubjectPolicy {
	return r.Spec.SubjectPolicy
}

// AccessPolicySpec defines the desired state of AccessPolicy
type AccessPolicySpec struct {
	SubjectPolicy `json:",inline"`
}

// AccessPolicyStatus defines the observed state of AccessPolicy.
type AccessPolicyStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// AccessPolicy is the Schema for the accesspolicies API
type AccessPolicy struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of AccessPolicy
	// +required
	Spec AccessPolicySpec `json:"spec"`

	// status defines the observed state of AccessPolicy
	// +optional
	Status AccessPolicyStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// AccessPolicyList contains a list of AccessPolicy
type AccessPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []AccessPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &AccessPolicy{}, &AccessPolicyList{})
		return nil
	})
}
