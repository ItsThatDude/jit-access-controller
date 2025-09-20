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

func (r *ClusterJITAccessRequest) GetSpec() *JITAccessRequestBaseSpec {
	return &r.Spec.JITAccessRequestBaseSpec
}
func (r *ClusterJITAccessRequest) GetStatus() *JITAccessRequestStatus {
	return &r.Status
}
func (r *ClusterJITAccessRequest) SetStatus(status *JITAccessRequestStatus) {
	r.Status = *status
}
func (r *ClusterJITAccessRequest) GetRoleKind() RoleKind {
	return RoleKindClusterRole
}
func (r *ClusterJITAccessRequest) GetScope() string {
	return "Cluster"
}
func (r *ClusterJITAccessRequest) GetNamespace() string {
	return ""
}
func (r *ClusterJITAccessRequest) GetName() string {
	return r.Name
}

// ClusterJITAccessRequestSpec defines the desired state of ClusterJITAccessRequest
type ClusterJITAccessRequestSpec struct {
	JITAccessRequestBaseSpec `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// ClusterJITAccessRequest is the Schema for the clusterjitaccessrequests API
type ClusterJITAccessRequest struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of ClusterJITAccessRequest
	// +required
	Spec ClusterJITAccessRequestSpec `json:"spec"`

	// status defines the observed state of ClusterJITAccessRequest
	// +optional
	Status JITAccessRequestStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// ClusterJITAccessRequestList contains a list of ClusterJITAccessRequest
type ClusterJITAccessRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterJITAccessRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterJITAccessRequest{}, &ClusterJITAccessRequestList{})
}
