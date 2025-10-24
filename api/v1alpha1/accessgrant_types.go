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

// AccessGrantSpec defines the desired state of AccessGrant
type AccessGrantSpec struct {
}

// +kubebuilder:validation:Enum=Cluster;Namespace
type GrantScope string

const (
	GrantScopeCluster   GrantScope = "Cluster"
	GrantScopeNamespace GrantScope = "Namespace"
)

// AccessGrantStatus defines the observed state of AccessGrant.
type AccessGrantStatus struct {
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

// AccessGrant is the Schema for the accessgrants API
type AccessGrant struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of AccessGrant
	// +optional
	Spec AccessGrantSpec `json:"spec,omitempty,omitzero"`

	// status defines the observed state of AccessGrant
	// +optional
	Status AccessGrantStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// AccessGrantList contains a list of AccessGrant
type AccessGrantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AccessGrant `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AccessGrant{}, &AccessGrantList{})
}
