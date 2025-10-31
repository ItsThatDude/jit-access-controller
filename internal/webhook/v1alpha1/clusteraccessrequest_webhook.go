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
	"net/http"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	accessv1alpha1 "antware.xyz/kairos/api/v1alpha1"
	"antware.xyz/kairos/internal/policy"
)

// +kubebuilder:webhook:path=/validate-access-antware-xyz-v1alpha1-clusteraccessrequest,mutating=false,failurePolicy=fail,sideEffects=None,groups=access.antware.xyz,resources=clusteraccessrequests,verbs=create;update,versions=v1alpha1,name=vclusteraccessrequest-v1alpha1.kb.io,admissionReviewVersions=v1

type ClusterAccessRequestValidator struct {
	decoder admission.Decoder
	client  client.Client
}

func SetupClusterAccessRequestWebhookWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(
		"/validate-access-antware-xyz-v1alpha1-clusteraccessrequest",
		&admission.Webhook{Handler: &ClusterAccessRequestValidator{
			decoder: admission.NewDecoder(mgr.GetScheme()),
			client:  mgr.GetClient(),
		}},
	)
}

func (v *ClusterAccessRequestValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &accessv1alpha1.ClusterAccessRequest{}

	if err := v.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if obj.Spec.Subject != req.UserInfo.Username {
		return admission.Denied("The subject must be the same as the user creating the request.")
	}

	if obj.Spec.Role.Name == "" && len(obj.Spec.Permissions) == 0 {
		return admission.Denied("either ClusterRole or Permissions needs to be set")
	}

	var policies accessv1alpha1.ClusterAccessPolicyList
	if err := v.client.List(ctx, &policies); err != nil {
		admission.Errored(http.StatusBadRequest, err)
	}

	permitted, _ := policy.IsRequestValid(obj, policies.Items)
	if !permitted {
		return admission.Denied("cluster access request did not match a policy")
	}

	return admission.Allowed("valid")
}
