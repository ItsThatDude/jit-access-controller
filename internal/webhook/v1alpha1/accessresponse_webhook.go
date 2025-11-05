package v1alpha1

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	accessv1alpha1 "antware.xyz/kairos/api/v1alpha1"
	"antware.xyz/kairos/internal/policy"
	"antware.xyz/kairos/internal/utils"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-access-antware-xyz-v1alpha1-accessresponse,mutating=false,failurePolicy=fail,sideEffects=None,groups=access.antware.xyz,resources=accessresponses,verbs=create;update,versions=v1alpha1,name=vaccessresponse-v1alpha1.kb.io,admissionReviewVersions=v1

type AccessResponseValidator struct {
	decoder        admission.Decoder
	client         client.Client
	namespace      string
	serviceAccount string
}

func SetupAccessResponseWebhookWithManager(mgr ctrl.Manager, namespace, serviceAccount string) {
	mgr.GetWebhookServer().Register(
		"/validate-access-antware-xyz-v1alpha1-accessresponse",
		&admission.Webhook{Handler: &AccessResponseValidator{
			decoder:        admission.NewDecoder(mgr.GetScheme()),
			client:         mgr.GetClient(),
			namespace:      namespace,
			serviceAccount: serviceAccount,
		}},
	)
}

func (v *AccessResponseValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &accessv1alpha1.AccessResponse{}

	if err := v.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if req.Operation == admissionv1.Delete {
		return admission.Allowed("deletion is allowed")
	}

	if req.Operation == admissionv1.Update {
		if req.UserInfo.Username == utils.FormatServiceAccountName(v.serviceAccount, v.namespace) {
			return admission.Allowed("kairos-controller-manager is allowed to update access requests")
		}
	}

	if req.Operation == admissionv1.Create && obj.Spec.Approver != req.UserInfo.Username {
		return admission.Denied("The approver must be the user creating the approval.")
	}

	request := &accessv1alpha1.AccessRequest{}
	if err := v.client.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: obj.Spec.RequestRef}, request); err != nil {
		return admission.Denied(fmt.Sprintf("an error occurred fetching the referenced AccessRequest: %s", err))
	}

	if request.Spec.Subject == obj.Spec.Approver {
		return admission.Denied("The approver can not be the same as the subject of the request.")
	}

	policies := &accessv1alpha1.AccessPolicyList{}
	if err := v.client.List(ctx, policies, client.InNamespace(req.Namespace)); err != nil {
		return admission.Denied(fmt.Sprintf("an error occurred fetching access policies in namespace '%s': %s", req.Namespace, err))
	}

	isRequestValid, matched_policy := policy.IsRequestValid(request, policies.Items)

	if !isRequestValid || matched_policy == nil {
		return admission.Denied(fmt.Sprintf("the request %s does not match an access policy in namespace '%s'", req.Name, req.Namespace))
	}

	switch req.Operation {
	case admissionv1.Create:
		if obj.Spec.Approver != req.UserInfo.Username {
			return admission.Denied(fmt.Sprintf("cannot specify an approver other than yourself (%s vs %s)", obj.Spec.Approver, req.UserInfo.Username))
		}

		group_matched := utils.SliceOverlaps(matched_policy.ApproverGroups, req.UserInfo.Groups)
		user_matched := slices.Contains(matched_policy.Approvers, req.UserInfo.Username)

		if !group_matched && !user_matched {
			return admission.Denied(fmt.Sprintf("user %s is not in the list of approvers for the matched policy", req.UserInfo.Username))
		}

	case admissionv1.Update:
		oldObj := &accessv1alpha1.AccessResponse{}
		if err := v.decoder.DecodeRaw(req.OldObject, oldObj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

	case admissionv1.Delete:
	}

	return admission.Allowed("valid")
}
