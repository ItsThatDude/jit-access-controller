package common

import (
	accessv1alpha1 "antware.xyz/jitaccess/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const JITFinalizer = "access.antware.xyz/finalizer"

type JITAccessRequestObject interface {
	client.Object
	GetSpec() *accessv1alpha1.JITAccessRequestBaseSpec
	GetStatus() *accessv1alpha1.JITAccessRequestStatus
	SetStatus(status *accessv1alpha1.JITAccessRequestStatus)
	GetScope() string
	GetNamespace() string
	GetName() string
}

type JITAccessResponseObject interface {
	client.Object
	GetResponse() accessv1alpha1.ResponseState
	GetApprover() string
}

type JITAccessPolicyListInterface interface {
	GetPolicies() []accessv1alpha1.SubjectPolicy
}
