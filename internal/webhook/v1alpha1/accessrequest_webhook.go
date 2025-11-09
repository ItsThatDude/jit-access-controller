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
	"reflect"

	admissionv1 "k8s.io/api/admission/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	accessv1alpha1 "github.com/itsthatdude/jit-access-controller/api/v1alpha1"
	"github.com/itsthatdude/jit-access-controller/internal/policy"
	"github.com/itsthatdude/jit-access-controller/internal/utils"
)

// +kubebuilder:webhook:path=/validate-access-antware-xyz-v1alpha1-accessrequest,mutating=false,failurePolicy=fail,sideEffects=None,groups=access.antware.xyz,resources=accessrequests,verbs=create;update,versions=v1alpha1,name=vaccessrequest-v1alpha1.kb.io,admissionReviewVersions=v1

type AccessRequestValidator struct {
	decoder        admission.Decoder
	client         client.Client
	namespace      string
	serviceAccount string
}

func SetupAccessRequestWebhookWithManager(mgr ctrl.Manager, namespace, serviceAccount string) {
	mgr.GetWebhookServer().Register(
		"/validate-access-antware-xyz-v1alpha1-accessrequest",
		&admission.Webhook{Handler: &AccessRequestValidator{
			decoder:        admission.NewDecoder(mgr.GetScheme()),
			client:         mgr.GetClient(),
			namespace:      namespace,
			serviceAccount: serviceAccount,
		}},
	)
}

func (v *AccessRequestValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &accessv1alpha1.AccessRequest{}

	if err := v.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if req.Operation == admissionv1.Delete {
		return admission.Allowed("deletion is allowed")
	}

	if req.Operation == admissionv1.Update {
		if req.UserInfo.Username == utils.FormatServiceAccountName(v.serviceAccount, v.namespace) {
			return admission.Allowed("jit-access-controller-manager is allowed to update access requests")
		}
	}

	if req.Operation == admissionv1.Create {
		if obj.Spec.Subject != req.UserInfo.Username {
			return admission.Denied("The subject must be the same as the user creating the request.")
		}
		if !reflect.DeepEqual(obj.Spec.Groups, req.UserInfo.Groups) {
			return admission.Denied("The subject's groups must be the same as the user creating the request.")
		}
	}

	if obj.Spec.Role.Name == "" && len(obj.Spec.Permissions) == 0 {
		return admission.Denied("either Role or Permissions needs to be set")
	}

	var policies accessv1alpha1.AccessPolicyList
	if err := v.client.List(ctx, &policies); err != nil {
		admission.Errored(http.StatusBadRequest, err)
	}

	permitted, _ := policy.IsRequestValid(obj, policies.Items)
	if !permitted {
		return admission.Denied("access request did not match a policy")
	}

	return admission.Allowed("valid")
}
