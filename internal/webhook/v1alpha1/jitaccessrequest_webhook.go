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
var jitaccessrequestlog = logf.Log.WithName("jitaccessrequest-resource")

// SetupJITAccessRequestWebhookWithManager registers the webhook for JITAccessRequest in the manager.
func SetupJITAccessRequestWebhookWithManager(mgr ctrl.Manager) error {
	validator := &JITAccessRequestCustomValidator{
		client: mgr.GetClient(),
	}
	return ctrl.NewWebhookManagedBy(mgr).For(&accessv1alpha1.JITAccessRequest{}).
		WithValidator(validator).
		WithDefaulter(&JITAccessRequestCustomDefaulter{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-access-antware-xyz-v1alpha1-jitaccessrequest,mutating=true,failurePolicy=fail,sideEffects=None,groups=access.antware.xyz,resources=jitaccessrequests,verbs=create;update,versions=v1alpha1,name=mjitaccessrequest-v1alpha1.kb.io,admissionReviewVersions=v1

// JITAccessRequestCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind JITAccessRequest when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type JITAccessRequestCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
}

var _ webhook.CustomDefaulter = &JITAccessRequestCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind JITAccessRequest.
func (d *JITAccessRequestCustomDefaulter) Default(_ context.Context, obj runtime.Object) error {
	jitaccessrequest, ok := obj.(*accessv1alpha1.JITAccessRequest)

	if !ok {
		return fmt.Errorf("expected an JITAccessRequest object but got %T", obj)
	}
	jitaccessrequestlog.Info("Defaulting for JITAccessRequest", "name", jitaccessrequest.GetName())

	// TODO(user): fill in your defaulting logic.

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-access-antware-xyz-v1alpha1-jitaccessrequest,mutating=false,failurePolicy=fail,sideEffects=None,groups=access.antware.xyz,resources=jitaccessrequests,verbs=create;update,versions=v1alpha1,name=vjitaccessrequest-v1alpha1.kb.io,admissionReviewVersions=v1

// JITAccessRequestCustomValidator struct is responsible for validating the JITAccessRequest resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type JITAccessRequestCustomValidator struct {
	client client.Client
}

var _ webhook.CustomValidator = &JITAccessRequestCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type JITAccessRequest.
func (v *JITAccessRequestCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	jitaccessrequest, ok := obj.(*accessv1alpha1.JITAccessRequest)
	if !ok {
		return nil, fmt.Errorf("expected a JITAccessRequest object but got %T", obj)
	}
	jitaccessrequestlog.Info("Validation for JITAccessRequest upon creation", "name", jitaccessrequest.GetName())

	var policies accessv1alpha1.JITAccessPolicyList
	if err := v.client.List(context.TODO(), &policies); err != nil {
		return nil, err
	}

	permitted := policy.ValidateNamespaced(jitaccessrequest, &policies)
	if !permitted {
		return nil, fmt.Errorf("access request did not match a policy")
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type JITAccessRequest.
func (v *JITAccessRequestCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	jitaccessrequest, ok := newObj.(*accessv1alpha1.JITAccessRequest)
	if !ok {
		return nil, fmt.Errorf("expected a JITAccessRequest object for the newObj but got %T", newObj)
	}
	jitaccessrequestlog.Info("Validation for JITAccessRequest upon update", "name", jitaccessrequest.GetName())

	// TODO(user): fill in your validation logic upon object update.

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type JITAccessRequest.
func (v *JITAccessRequestCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	jitaccessrequest, ok := obj.(*accessv1alpha1.JITAccessRequest)
	if !ok {
		return nil, fmt.Errorf("expected a JITAccessRequest object but got %T", obj)
	}
	jitaccessrequestlog.Info("Validation for JITAccessRequest upon deletion", "name", jitaccessrequest.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}
