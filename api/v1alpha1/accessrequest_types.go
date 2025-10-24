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

func (r *AccessRequest) GetSpec() *AccessRequestBaseSpec {
	return &r.Spec.AccessRequestBaseSpec
}
func (r *AccessRequest) GetStatus() *AccessRequestStatus {
	return &r.Status
}
func (r *AccessRequest) SetStatus(status *AccessRequestStatus) {
	r.Status = *status
}
func (r *AccessRequest) GetScope() string {
	return "Namespace"
}
func (r *AccessRequest) GetNamespace() string {
	return r.Namespace
}
func (r *AccessRequest) GetName() string {
	return r.Name
}

// AccessRequestSpec defines the desired state of AccessRequest
type AccessRequestSpec struct {
	AccessRequestBaseSpec `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// AccessRequest is the Schema for the accessrequests API
type AccessRequest struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of AccessRequest
	// +required
	Spec AccessRequestSpec `json:"spec"`

	// status defines the observed state of AccessRequest
	// +optional
	Status AccessRequestStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// AccessRequestList contains a list of AccessRequest
type AccessRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AccessRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AccessRequest{}, &AccessRequestList{})
}
