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

func (r *JITAccessRequest) GetSpec() *JITAccessRequestBaseSpec {
	return &r.Spec.JITAccessRequestBaseSpec
}
func (r *JITAccessRequest) GetStatus() *JITAccessRequestStatus {
	return &r.Status
}
func (r *JITAccessRequest) SetStatus(st *JITAccessRequestStatus) {
	r.Status = *st
}
func (r *JITAccessRequest) GetRoleKind() RoleKind {
	return r.Spec.RoleKind
}

// JITAccessRequestSpec defines the desired state of JITAccessRequest
type JITAccessRequestSpec struct {
	JITAccessRequestBaseSpec `json:",inline"`

	// Type of Role - Role or ClusterRole
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="RoleKind cannot be changed after creation"
	RoleKind RoleKind `json:"roleKind,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// JITAccessRequest is the Schema for the jitaccessrequests API
type JITAccessRequest struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of JITAccessRequest
	// +required
	Spec JITAccessRequestSpec `json:"spec"`

	// status defines the observed state of JITAccessRequest
	// +optional
	Status JITAccessRequestStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// JITAccessRequestList contains a list of JITAccessRequest
type JITAccessRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JITAccessRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&JITAccessRequest{}, &JITAccessRequestList{})
}
