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
	SchemeBuilder.Register(&ClusterAccessGrant{}, &ClusterAccessGrantList{})
}
