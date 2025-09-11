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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// JITAccessRequestSpec defines the desired state of JITAccessRequest
type JITAccessRequestSpec struct {
	Subject string `json:"subject"`

	// Name of the Role
	// +optional
	Role string `json:"role,omitempty"`

	// Type of Role - Role or ClusterRole
	// +optional
	RoleKind RoleKind `json:"roleKind,omitempty"`

	// A list of adhoc permissions to request
	// +optional
	Permissions []rbacv1.PolicyRule `json:"permissions,omitempty"`

	// Duration in seconds (e.g. 600 for 10 min)
	DurationSeconds int64 `json:"durationSeconds"`

	// User's justification for the request
	Justification string `json:"justification"`
}

type RoleKind string

const (
	RoleKindRole        RoleKind = "Role"
	RoleKindClusterRole RoleKind = "ClusterRole"
)

// JITAccessRequestStatus defines the observed state of JITAccessRequest.
type JITAccessRequestStatus struct {
	// ID of the access request
	RequestId string `json:"requestId"`

	// State of the Access Request
	State RequestState `json:"state,omitempty"`

	// Timestamp when the access will expire
	ExpiresAt *metav1.Time `json:"expiresAt,omitempty"`

	// True/False if the RoleBinding has been created
	RoleBindingCreated bool `json:"roleBindingCreated"`

	// True/False if the Adhoc Role has been created
	AdhocRoleCreated bool `json:"adhocRoleCreated"`

	// True/False if the Adhoc Role has been created
	AdhocRoleBindingCreated bool `json:"adhocRoleBindingCreated"`
}

type RequestState string

const (
	RequestStatePending  RequestState = "Pending"
	RequestStateApproved RequestState = "Approved"
	RequestStateDenied   RequestState = "Denied"
)

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
