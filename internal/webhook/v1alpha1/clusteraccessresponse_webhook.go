package v1alpha1

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	accessv1alpha1 "github.com/itsthatdude/jit-access-controller/api/v1alpha1"
	"github.com/itsthatdude/jit-access-controller/internal/policy"
	"github.com/itsthatdude/jit-access-controller/internal/utils"
	admissionv1 "k8s.io/api/admission/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-access-antware-xyz-v1alpha1-clusteraccessresponse,mutating=false,failurePolicy=fail,sideEffects=None,groups=access.antware.xyz,resources=clusteraccessresponses,verbs=create;update,versions=v1alpha1,name=vclusteraccessresponse-v1alpha1.kb.io,admissionReviewVersions=v1

type ClusterAccessResponseValidator struct {
	decoder                admission.Decoder
	client                 client.Client
	namespace              string
	serviceAccount         string
	frontendServiceAccount string
	PolicyManager          *policy.PolicyManager
	PolicyResolver         *policy.PolicyResolver
}

func SetupClusterAccessResponseWebhookWithManager(mgr ctrl.Manager, namespace, serviceAccount string, frontendServiceAccount string, policyManager *policy.PolicyManager) {
	mgr.GetWebhookServer().Register(
		"/validate-access-antware-xyz-v1alpha1-clusteraccessresponse",
		&admission.Webhook{Handler: &ClusterAccessResponseValidator{
			decoder:                admission.NewDecoder(mgr.GetScheme()),
			client:                 mgr.GetClient(),
			namespace:              namespace,
			serviceAccount:         serviceAccount,
			frontendServiceAccount: frontendServiceAccount,
			PolicyManager:          policyManager,
			PolicyResolver:         &policy.PolicyResolver{},
		}},
	)
}

func (v *ClusterAccessResponseValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &accessv1alpha1.ClusterAccessResponse{}
	isController := utils.IsController(v.namespace, v.serviceAccount, req.UserInfo)
	isFrontend := utils.IsController(v.namespace, v.frontendServiceAccount, req.UserInfo)

	if err := v.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if req.Operation == admissionv1.Delete {
		return admission.Allowed("deletion is allowed")
	}

	if req.Operation == admissionv1.Update && isController {
		return admission.Allowed("jit-access-controller-manager is allowed to update access requests")
	}

	if !isFrontend && req.Operation == admissionv1.Create && obj.Spec.Approver != req.UserInfo.Username {
		return admission.Denied("The approver must be the same as the user creating the approval.")
	}

	request := &accessv1alpha1.ClusterAccessRequest{}
	if err := v.client.Get(ctx, types.NamespacedName{Name: obj.Spec.RequestRef}, request); err != nil {
		return admission.Denied(fmt.Sprintf("an error occurred fetching the referenced ClusterAccessRequest: %s", err))
	}

	if request.Spec.Subject == obj.Spec.Approver {
		return admission.Denied("The approver can not be the same as the subject of the request.")
	}

	policies := v.PolicyManager.GetSnapshot()
	matched_policy := v.PolicyResolver.Resolve(request, policies)

	if matched_policy == nil {
		return admission.Denied(fmt.Sprintf("the request %s does not match an access policy", req.Name))
	}

	policySpec := matched_policy.GetPolicy()

	switch req.Operation {
	case admissionv1.Create:
		if !isFrontend && obj.Spec.Approver != req.UserInfo.Username {
			return admission.Denied(fmt.Sprintf("cannot specify an approver other than yourself (%s vs %s)", obj.Spec.Approver, req.UserInfo.Username))
		}

		userNames := []string{}
		groupNames := []string{}
		for _, approver := range policySpec.Approvers {
			switch approver.Kind {
			case rbacv1.UserKind:
				userNames = append(userNames, approver.Name)
			case rbacv1.GroupKind:
				groupNames = append(groupNames, approver.Name)
			}
		}

		group_matched := utils.SliceOverlaps(groupNames, req.UserInfo.Groups)
		user_matched := slices.Contains(userNames, req.UserInfo.Username)

		if !group_matched && !user_matched && !isFrontend {
			return admission.Denied(fmt.Sprintf("user %s is not in the list of approvers for the matched policy", req.UserInfo.Username))
		}

	case admissionv1.Update:
		oldObj := &accessv1alpha1.ClusterAccessResponse{}
		if err := v.decoder.DecodeRaw(req.OldObject, oldObj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

	case admissionv1.Delete:
	}

	return admission.Allowed("valid")
}
