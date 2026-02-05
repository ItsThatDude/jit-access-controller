package common

import (
	"github.com/itsthatdude/jit-access-controller/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const JITFinalizer = "access.antware.xyz/finalizer"

type AccessGrantObject interface {
	client.Object
	GetStatus() *v1alpha1.AccessGrantStatus
	SetStatus(status *v1alpha1.AccessGrantStatus)
	GetScope() v1alpha1.RequestScope
	GetName() string
}

type AccessRequestObject interface {
	client.Object
	GetSpec() *v1alpha1.AccessRequestBaseSpec
	GetStatus() *v1alpha1.AccessRequestStatus
	SetStatus(status *v1alpha1.AccessRequestStatus)
	GetScope() v1alpha1.RequestScope
	GetNamespace() string
	GetName() string
	GetSubject() string
}

type AccessResponseObject interface {
	client.Object
	GetResponse() v1alpha1.ResponseState
	GetApprover() string
}

type AccessPolicyObject interface {
	GetName() string
	GetNamespace() string
	GetScope() v1alpha1.PolicyScope
	GetPolicy() v1alpha1.SubjectPolicy
}
