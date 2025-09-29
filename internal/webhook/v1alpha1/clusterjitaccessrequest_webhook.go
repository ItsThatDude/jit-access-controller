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
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	accessv1alpha1 "antware.xyz/jitaccess/api/v1alpha1"
	"antware.xyz/jitaccess/internal/policy"
)

// nolint:unused
// log is for logging in this package.
var clusterjitaccessrequestlog = logf.Log.WithName("clusterjitaccessrequest-resource")

// SetupClusterJITAccessRequestWebhookWithManager registers the webhook for ClusterJITAccessRequest in the manager.
func SetupClusterJITAccessRequestWebhookWithManager(mgr ctrl.Manager) error {
	validator := &ClusterJITAccessRequestCustomValidator{
		client: mgr.GetClient(),
	}
	return ctrl.NewWebhookManagedBy(mgr).For(&accessv1alpha1.ClusterJITAccessRequest{}).
		WithValidator(validator).
		WithDefaulter(&ClusterJITAccessRequestCustomDefaulter{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-access-antware-xyz-v1alpha1-clusterjitaccessrequest,mutating=true,failurePolicy=fail,sideEffects=None,groups=access.antware.xyz,resources=clusterjitaccessrequests,verbs=create;update,versions=v1alpha1,name=mclusterjitaccessrequest-v1alpha1.kb.io,admissionReviewVersions=v1

// ClusterJITAccessRequestCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind ClusterJITAccessRequest when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type ClusterJITAccessRequestCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
}

var _ webhook.CustomDefaulter = &ClusterJITAccessRequestCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind ClusterJITAccessRequest.
func (d *ClusterJITAccessRequestCustomDefaulter) Default(_ context.Context, obj runtime.Object) error {
	req, ok := obj.(*accessv1alpha1.ClusterJITAccessRequest)

	if !ok {
		return fmt.Errorf("expected an ClusterJITAccessRequest object but got %T", obj)
	}
	clusterjitaccessrequestlog.Info("Defaulting for ClusterJITAccessRequest", "name", req.GetName())

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-access-antware-xyz-v1alpha1-clusterjitaccessrequest,mutating=false,failurePolicy=fail,sideEffects=None,groups=access.antware.xyz,resources=clusterjitaccessrequests,verbs=create;update,versions=v1alpha1,name=vclusterjitaccessrequest-v1alpha1.kb.io,admissionReviewVersions=v1

// ClusterJITAccessRequestCustomValidator struct is responsible for validating the ClusterJITAccessRequest resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type ClusterJITAccessRequestCustomValidator struct {
	client client.Client
}

var _ webhook.CustomValidator = &ClusterJITAccessRequestCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type ClusterJITAccessRequest.
func (v *ClusterJITAccessRequestCustomValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	clusterjitaccessrequest, ok := obj.(*accessv1alpha1.ClusterJITAccessRequest)
	if !ok {
		return nil, fmt.Errorf("expected a ClusterJITAccessRequest object but got %T", obj)
	}

	if clusterjitaccessrequest.Spec.Role == "" && len(clusterjitaccessrequest.Spec.Permissions) == 0 {
		return nil, fmt.Errorf("either ClusterRole or Permissions needs to be set")
	}

	var policies accessv1alpha1.ClusterJITAccessPolicyList
	if err := v.client.List(context.TODO(), &policies); err != nil {
		return nil, err
	}

	permitted, _ := policy.IsRequestValid(clusterjitaccessrequest, policies.Items)
	if !permitted {
		return nil, fmt.Errorf("cluster access request did not match a policy")
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type ClusterJITAccessRequest.
func (v *ClusterJITAccessRequestCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	/*
		clusterjitaccessrequest, ok := newObj.(*accessv1alpha1.ClusterJITAccessRequest)
		if !ok {
			return nil, fmt.Errorf("expected a ClusterJITAccessRequest object for the newObj but got %T", newObj)
		}
		clusterjitaccessrequestlog.Info("Validation for ClusterJITAccessRequest upon update", "name", clusterjitaccessrequest.GetName())
	*/

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type ClusterJITAccessRequest.
func (v *ClusterJITAccessRequestCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	/*
		clusterjitaccessrequest, ok := obj.(*accessv1alpha1.ClusterJITAccessRequest)
		if !ok {
			return nil, fmt.Errorf("expected a ClusterJITAccessRequest object but got %T", obj)
		}
		clusterjitaccessrequestlog.Info("Validation for ClusterJITAccessRequest upon deletion", "name", clusterjitaccessrequest.GetName())
	*/

	return nil, nil
}
