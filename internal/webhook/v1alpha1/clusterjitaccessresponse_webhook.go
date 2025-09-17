package v1alpha1

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	accessv1alpha1 "antware.xyz/jitaccess/api/v1alpha1"
	"antware.xyz/jitaccess/internal/policy"
	"antware.xyz/jitaccess/internal/utils"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-access-antware-xyz-v1alpha1-clusterjitaccessresponse,mutating=false,failurePolicy=fail,sideEffects=None,groups=access.antware.xyz,resources=clusterjitaccessresponses,verbs=create;update,versions=v1alpha1,name=vclusterjitaccessresponse-v1alpha1.kb.io,admissionReviewVersions=v1

type ClusterJITAccessResponseValidator struct {
	decoder admission.Decoder
	client  client.Client
}

func SetupClusterJITAccessResponseWebhookWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(
		"/validate-access-antware-xyz-v1alpha1-clusterjitaccessresponse",
		&admission.Webhook{Handler: &JITAccessResponseValidator{
			decoder: admission.NewDecoder(mgr.GetScheme()),
			client:  mgr.GetClient(),
		}},
	)
}

func (v *ClusterJITAccessResponseValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &accessv1alpha1.ClusterJITAccessResponse{}

	if err := v.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	request := &accessv1alpha1.ClusterJITAccessRequest{}
	if err := v.client.Get(ctx, types.NamespacedName{Name: obj.Spec.RequestRef}, request); err != nil {
		return admission.Denied(fmt.Sprintf("an error occured fetching the referenced ClusterJITAccessRequest: %s", err))
	}

	policies := &accessv1alpha1.ClusterJITAccessPolicyList{}
	if err := v.client.List(ctx, policies); err != nil {
		return admission.Denied(fmt.Sprintf("an error occured fetching access policies: %s", err))
	}

	isRequestValid, matched_policy := policy.IsClusterRequestValid(request, policies)

	if !isRequestValid || matched_policy == nil {
		return admission.Denied(fmt.Sprintf("the request %s does not match an access policy", req.Name))
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
		oldObj := &accessv1alpha1.ClusterJITAccessResponse{}
		if err := v.decoder.DecodeRaw(req.AdmissionRequest.OldObject, oldObj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

	case admissionv1.Delete:
	}

	return admission.Allowed("valid")
}
