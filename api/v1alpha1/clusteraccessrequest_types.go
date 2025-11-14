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

func (r *ClusterAccessRequest) GetSpec() *AccessRequestBaseSpec {
	return &r.Spec.AccessRequestBaseSpec
}
func (r *ClusterAccessRequest) GetStatus() *AccessRequestStatus {
	return &r.Status
}
func (r *ClusterAccessRequest) SetStatus(status *AccessRequestStatus) {
	r.Status = *status
}
func (r *ClusterAccessRequest) GetScope() RequestScope {
	return RequestScopeCluster
}
func (r *ClusterAccessRequest) GetNamespace() string {
	return ""
}
func (r *ClusterAccessRequest) GetName() string {
	return r.Name
}
func (r *ClusterAccessRequest) GetSubject() string {
	return r.Spec.Subject
}

// ClusterAccessRequestSpec defines the desired state of ClusterAccessRequest
type ClusterAccessRequestSpec struct {
	AccessRequestBaseSpec `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:selectablefield:JSONPath=".status.state"

// ClusterAccessRequest is the Schema for the clusteraccessrequests API
// +kubebuilder:printcolumn:name="Approvals-Required",type=integer,JSONPath=`.status.approvalsRequired`
// +kubebuilder:printcolumn:name="Request-Expires-At",type=date,JSONPath=`.status.requestExpiresAt`
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type ClusterAccessRequest struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of ClusterAccessRequest
	// +required
	Spec ClusterAccessRequestSpec `json:"spec"`

	// status defines the observed state of ClusterAccessRequest
	// +optional
	Status AccessRequestStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// ClusterAccessRequestList contains a list of ClusterAccessRequest
type ClusterAccessRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterAccessRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterAccessRequest{}, &ClusterAccessRequestList{})
}
