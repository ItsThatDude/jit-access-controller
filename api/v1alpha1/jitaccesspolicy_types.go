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

// JITAccessPolicySpec defines the desired state of JITAccessPolicy
type JITAccessPolicySpec struct {
	Policies []SubjectPolicy `json:"policies"`
}

// SubjectPolicy defines access rules for a single subject (user/serviceaccount).
type SubjectPolicy struct {
	// Subject is the identity (email or K8s username) of the requester.
	Subjects []string `json:"subjects"`

	// AllowedRoles is a list of roles the subject is allowed to request.
	AllowedRoles []string `json:"allowedRoles,omitempty"`

	// AllowedPermissions is a list of adhoc permissions the subject is allowed to request.
	AllowedPermissions []rbacv1.PolicyRule `json:"allowedPermissions,omitempty"`

	// MaxDurationSeconds is the max duration for temporary access.
	// +kubebuilder:validation:Minimum=1
	MaxDurationSeconds int64 `json:"maxDurationSeconds"`

	// The minimum number of approvals required to grant the request
	// +kubebuilder:validation:Minimum=1
	RequiredApprovals int `json:"requiredApprovals"`

	// Approvers
	Approvers []string `json:"approvers,omitempty"`

	// Approver Groups
	ApproverGroups []string `json:"approverGroups,omitempty"`
}

// JITAccessPolicyStatus defines the observed state of JITAccessPolicy.
type JITAccessPolicyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// JITAccessPolicy is the Schema for the jitaccesspolicies API
type JITAccessPolicy struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of JITAccessPolicy
	// +required
	Spec JITAccessPolicySpec `json:"spec"`

	// status defines the observed state of JITAccessPolicy
	// +optional
	Status JITAccessPolicyStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// JITAccessPolicyList contains a list of JITAccessPolicy
type JITAccessPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JITAccessPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&JITAccessPolicy{}, &JITAccessPolicyList{})
}
