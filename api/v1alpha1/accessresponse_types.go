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

func (r *AccessResponse) GetResponse() ResponseState {
	return r.Spec.Response
}

func (r *AccessResponse) GetApprover() string {
	return r.Spec.Approver
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// AccessResponse is the Schema for the accessresponses API
type AccessResponse struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of AccessResponse
	// +required
	Spec AccessResponseSpec `json:"spec"`

	// status defines the observed state of AccessResponse
	// +optional
	Status AccessResponseStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// AccessResponseList contains a list of AccessResponse
type AccessResponseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AccessResponse `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AccessResponse{}, &AccessResponseList{})
}
