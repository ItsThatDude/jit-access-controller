/*
Copyright 2026.

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

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	accessv1alpha1 "github.com/itsthatdude/jit-access-controller/api/v1alpha1"
)

// nolint:unused
// log is for logging in this package.
var accessrequestlog = logf.Log.WithName("accessrequest-resource")

// SetupAccessRequestWebhookWithManager registers the webhook for AccessRequest in the manager.
func SetupAccessRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &accessv1alpha1.AccessRequest{}).
		WithValidator(&AccessRequestCustomValidator{}).
		WithDefaulter(&AccessRequestCustomDefaulter{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-access-antware-xyz-v1alpha1-accessrequest,mutating=true,failurePolicy=fail,sideEffects=None,groups=access.antware.xyz,resources=accessrequests,verbs=create;update,versions=v1alpha1,name=maccessrequest-v1alpha1.kb.io,admissionReviewVersions=v1

// AccessRequestCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind AccessRequest when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type AccessRequestCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind AccessRequest.
func (d *AccessRequestCustomDefaulter) Default(_ context.Context, obj *accessv1alpha1.AccessRequest) error {
	accessrequestlog.Info("Defaulting for AccessRequest", "name", obj.GetName())

	// TODO(user): fill in your defaulting logic.

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: If you want to customise the 'path', use the flags '--defaulting-path' or '--validation-path'.
// +kubebuilder:webhook:path=/validate-access-antware-xyz-v1alpha1-accessrequest,mutating=false,failurePolicy=fail,sideEffects=None,groups=access.antware.xyz,resources=accessrequests,verbs=create;update,versions=v1alpha1,name=vaccessrequest-v1alpha1.kb.io,admissionReviewVersions=v1

// AccessRequestCustomValidator struct is responsible for validating the AccessRequest resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type AccessRequestCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type AccessRequest.
func (v *AccessRequestCustomValidator) ValidateCreate(_ context.Context, obj *accessv1alpha1.AccessRequest) (admission.Warnings, error) {
	accessrequestlog.Info("Validation for AccessRequest upon creation", "name", obj.GetName())

	// TODO(user): fill in your validation logic upon object creation.

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type AccessRequest.
func (v *AccessRequestCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj *accessv1alpha1.AccessRequest) (admission.Warnings, error) {
	accessrequestlog.Info("Validation for AccessRequest upon update", "name", newObj.GetName())

	// TODO(user): fill in your validation logic upon object update.

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type AccessRequest.
func (v *AccessRequestCustomValidator) ValidateDelete(_ context.Context, obj *accessv1alpha1.AccessRequest) (admission.Warnings, error) {
	accessrequestlog.Info("Validation for AccessRequest upon deletion", "name", obj.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}
