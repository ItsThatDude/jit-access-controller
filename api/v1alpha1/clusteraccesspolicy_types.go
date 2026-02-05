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
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of ClusterAccessPolicy
	// +required
	Spec ClusterAccessPolicySpec `json:"spec"`

	// status defines the observed state of ClusterAccessPolicy
	// +optional
	Status AccessPolicyStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// ClusterAccessPolicyList contains a list of ClusterAccessPolicy
type ClusterAccessPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterAccessPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterAccessPolicy{}, &ClusterAccessPolicyList{})
}
