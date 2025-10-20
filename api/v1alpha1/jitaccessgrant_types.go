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

// JITAccessGrantSpec defines the desired state of JITAccessGrant
type JITAccessGrantSpec struct {
}

// +kubebuilder:validation:Enum=Cluster;Namespace
type GrantScope string

const (
	GrantScopeCluster   GrantScope = "Cluster"
	GrantScopeNamespace GrantScope = "Namespace"
)

// JITAccessGrantStatus defines the observed state of JITAccessGrant.
type JITAccessGrantStatus struct {
	Request   string `json:"request"`
	RequestId string `json:"requestId"`

	Subject    string   `json:"subject"`
	ApprovedBy []string `json:"approvedBy"`

	Scope       GrantScope          `json:"scope,omitempty"`
	Namespace   string              `json:"namespace,omitempty"`
	Role        rbacv1.RoleRef      `json:"role,omitempty"`
	Permissions []rbacv1.PolicyRule `json:"permissions,omitempty"`
	Duration    string              `json:"duration"`

	AccessExpiresAt         *metav1.Time `json:"accessExpiresAt,omitempty"`
	RoleBindingCreated      bool         `json:"roleBindingCreated,omitempty"`
	AdhocRoleCreated        bool         `json:"adhocRoleCreated,omitempty"`
	AdhocRoleBindingCreated bool         `json:"adhocRoleBindingCreated,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// JITAccessGrant is the Schema for the jitaccessgrants API
type JITAccessGrant struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of JITAccessGrant
	// +optional
	Spec JITAccessGrantSpec `json:"spec,omitempty,omitzero"`

	// status defines the observed state of JITAccessGrant
	// +optional
	Status JITAccessGrantStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// JITAccessGrantList contains a list of JITAccessGrant
type JITAccessGrantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JITAccessGrant `json:"items"`
}

func init() {
	SchemeBuilder.Register(&JITAccessGrant{}, &JITAccessGrantList{})
}
