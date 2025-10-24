package common

import (
	accessv1alpha1 "antware.xyz/jitaccess/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const JITFinalizer = "access.antware.xyz/finalizer"

type AccessRequestObject interface {
	client.Object
	GetSpec() *accessv1alpha1.AccessRequestBaseSpec
	GetStatus() *accessv1alpha1.AccessRequestStatus
	SetStatus(status *accessv1alpha1.AccessRequestStatus)
	GetScope() string
	GetNamespace() string
	GetName() string
}

type AccessResponseObject interface {
	client.Object
	GetResponse() accessv1alpha1.ResponseState
	GetApprover() string
}

type AccessPolicyListInterface interface {
	GetPolicy() accessv1alpha1.SubjectPolicy
}
