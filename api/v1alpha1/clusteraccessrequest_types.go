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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterAccessRequestSpec defines the desired state of ClusterAccessRequest
type ClusterAccessRequestSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// foo is an example field of ClusterAccessRequest. Edit clusteraccessrequest_types.go to remove/update
	// +optional
	Foo *string `json:"foo,omitempty"`
}

// ClusterAccessRequestStatus defines the observed state of ClusterAccessRequest.
type ClusterAccessRequestStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// ClusterAccessRequest is the Schema for the clusteraccessrequests API
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
	Status ClusterAccessRequestStatus `json:"status,omitempty,omitzero"`
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
