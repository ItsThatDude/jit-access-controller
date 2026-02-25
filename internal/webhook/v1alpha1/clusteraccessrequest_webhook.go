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
var clusteraccessrequestlog = logf.Log.WithName("clusteraccessrequest-resource")

// SetupClusterAccessRequestWebhookWithManager registers the webhook for ClusterAccessRequest in the manager.
func SetupClusterAccessRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &accessv1alpha1.ClusterAccessRequest{}).
		WithValidator(&ClusterAccessRequestCustomValidator{}).
		WithDefaulter(&ClusterAccessRequestCustomDefaulter{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-access-antware-xyz-v1alpha1-clusteraccessrequest,mutating=true,failurePolicy=fail,sideEffects=None,groups=access.antware.xyz,resources=clusteraccessrequests,verbs=create;update,versions=v1alpha1,name=mclusteraccessrequest-v1alpha1.kb.io,admissionReviewVersions=v1

// ClusterAccessRequestCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind ClusterAccessRequest when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type ClusterAccessRequestCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind ClusterAccessRequest.
func (d *ClusterAccessRequestCustomDefaulter) Default(_ context.Context, obj *accessv1alpha1.ClusterAccessRequest) error {
	clusteraccessrequestlog.Info("Defaulting for ClusterAccessRequest", "name", obj.GetName())

	// TODO(user): fill in your defaulting logic.

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: If you want to customise the 'path', use the flags '--defaulting-path' or '--validation-path'.
// +kubebuilder:webhook:path=/validate-access-antware-xyz-v1alpha1-clusteraccessrequest,mutating=false,failurePolicy=fail,sideEffects=None,groups=access.antware.xyz,resources=clusteraccessrequests,verbs=create;update,versions=v1alpha1,name=vclusteraccessrequest-v1alpha1.kb.io,admissionReviewVersions=v1

// ClusterAccessRequestCustomValidator struct is responsible for validating the ClusterAccessRequest resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type ClusterAccessRequestCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type ClusterAccessRequest.
func (v *ClusterAccessRequestCustomValidator) ValidateCreate(_ context.Context, obj *accessv1alpha1.ClusterAccessRequest) (admission.Warnings, error) {
	clusteraccessrequestlog.Info("Validation for ClusterAccessRequest upon creation", "name", obj.GetName())

	// TODO(user): fill in your validation logic upon object creation.

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type ClusterAccessRequest.
func (v *ClusterAccessRequestCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj *accessv1alpha1.ClusterAccessRequest) (admission.Warnings, error) {
	clusteraccessrequestlog.Info("Validation for ClusterAccessRequest upon update", "name", newObj.GetName())

	// TODO(user): fill in your validation logic upon object update.

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type ClusterAccessRequest.
func (v *ClusterAccessRequestCustomValidator) ValidateDelete(_ context.Context, obj *accessv1alpha1.ClusterAccessRequest) (admission.Warnings, error) {
	clusteraccessrequestlog.Info("Validation for ClusterAccessRequest upon deletion", "name", obj.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}
