/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r ClusterJITAccessPolicy) GetPolicy() SubjectPolicy {
	return r.Spec.SubjectPolicy
}

// ClusterJITAccessPolicySpec defines the desired state of ClusterJITAccessPolicy
type ClusterJITAccessPolicySpec struct {
	SubjectPolicy `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// ClusterJITAccessPolicy is the Schema for the clusterjitaccesspolicies API
type ClusterJITAccessPolicy struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of ClusterJITAccessPolicy
	// +required
	Spec ClusterJITAccessPolicySpec `json:"spec"`

	// status defines the observed state of ClusterJITAccessPolicy
	// +optional
	Status JITAccessPolicyStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// ClusterJITAccessPolicyList contains a list of ClusterJITAccessPolicy
type ClusterJITAccessPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterJITAccessPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterJITAccessPolicy{}, &ClusterJITAccessPolicyList{})
}
