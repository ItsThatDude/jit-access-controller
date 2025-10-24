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
var clusteraccessrequestlog = logf.Log.WithName("clusteraccessrequest-resource")

// SetupClusterAccessRequestWebhookWithManager registers the webhook for ClusterAccessRequest in the manager.
func SetupClusterAccessRequestWebhookWithManager(mgr ctrl.Manager) error {
	validator := &ClusterAccessRequestCustomValidator{
		client: mgr.GetClient(),
	}
	return ctrl.NewWebhookManagedBy(mgr).For(&accessv1alpha1.ClusterAccessRequest{}).
		WithValidator(validator).
		Complete()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-access-antware-xyz-v1alpha1-clusteraccessrequest,mutating=false,failurePolicy=fail,sideEffects=None,groups=access.antware.xyz,resources=clusteraccessrequests,verbs=create;update,versions=v1alpha1,name=vclusteraccessrequest-v1alpha1.kb.io,admissionReviewVersions=v1

// ClusterAccessRequestCustomValidator struct is responsible for validating the ClusterAccessRequest resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type ClusterAccessRequestCustomValidator struct {
	client client.Client
}

var _ webhook.CustomValidator = &ClusterAccessRequestCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type ClusterAccessRequest.
func (v *ClusterAccessRequestCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	clusteraccessrequest, ok := obj.(*accessv1alpha1.ClusterAccessRequest)
	if !ok {
		return nil, fmt.Errorf("expected a ClusterAccessRequest object but got %T", obj)
	}

	if clusteraccessrequest.Spec.Role.Name == "" && len(clusteraccessrequest.Spec.Permissions) == 0 {
		return nil, fmt.Errorf("either ClusterRole or Permissions needs to be set")
	}

	var policies accessv1alpha1.ClusterAccessPolicyList
	if err := v.client.List(ctx, &policies); err != nil {
		return nil, err
	}

	permitted, _ := policy.IsRequestValid(clusteraccessrequest, policies.Items)
	if !permitted {
		return nil, fmt.Errorf("cluster access request did not match a policy")
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type ClusterAccessRequest.
func (v *ClusterAccessRequestCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	/*
		clusteraccessrequest, ok := newObj.(*accessv1alpha1.ClusterAccessRequest)
		if !ok {
			return nil, fmt.Errorf("expected a ClusterAccessRequest object for the newObj but got %T", newObj)
		}
		clusteraccessrequestlog.Info("Validation for ClusterAccessRequest upon update", "name", clusteraccessrequest.GetName())
	*/

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type ClusterAccessRequest.
func (v *ClusterAccessRequestCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	/*
		clusteraccessrequest, ok := obj.(*accessv1alpha1.ClusterAccessRequest)
		if !ok {
			return nil, fmt.Errorf("expected a ClusterAccessRequest object but got %T", obj)
		}
		clusteraccessrequestlog.Info("Validation for ClusterAccessRequest upon deletion", "name", clusteraccessrequest.GetName())
	*/

	return nil, nil
}
