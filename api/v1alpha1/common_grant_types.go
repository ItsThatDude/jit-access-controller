package v1alpha1

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AccessGrantStatus struct {
	Request   string `json:"request"`
	RequestId string `json:"requestId"`

	Subject    string   `json:"subject"`
	ApprovedBy []string `json:"approvedBy"`

	Role        rbacv1.RoleRef      `json:"role,omitempty"`
	Permissions []rbacv1.PolicyRule `json:"permissions,omitempty"`
	Duration    string              `json:"duration"`

	AccessExpiresAt         *metav1.Time `json:"accessExpiresAt,omitempty"`
	RoleBindingCreated      bool         `json:"roleBindingCreated,omitempty"`
	AdhocRoleCreated        bool         `json:"adhocRoleCreated,omitempty"`
	AdhocRoleBindingCreated bool         `json:"adhocRoleBindingCreated,omitempty"`
}
