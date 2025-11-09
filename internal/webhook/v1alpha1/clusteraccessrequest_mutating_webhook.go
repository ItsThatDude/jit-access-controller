package v1alpha1

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/itsthatdude/jit-access-controller/api/v1alpha1"
	admissionv1 "k8s.io/api/admission/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/mutate-access-antware-xyz-v1alpha1-clusteraccessrequest,mutating=true,failurePolicy=fail,sideEffects=None,groups=access.antware.xyz,resources=clusteraccessrequests,verbs=create;update,versions=v1alpha1,name=mclusteraccessrequest-v1alpha1.kb.io,admissionReviewVersions=v1

type ClusterAccessRequestMutator struct {
	decoder admission.Decoder
}

func SetupClusterAccessRequestMutatingWebhookWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(
		"/mutate-access-antware-xyz-v1alpha1-clusteraccessrequest",
		&admission.Webhook{Handler: &ClusterAccessRequestMutator{
			decoder: admission.NewDecoder(mgr.GetScheme()),
		}},
	)
}

func (m *ClusterAccessRequestMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &v1alpha1.ClusterAccessRequest{}

	if err := m.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if req.Operation == admissionv1.Create {
		obj.Spec.Subject = req.UserInfo.Username
		obj.Spec.Groups = req.UserInfo.Groups
	}

	marshaled, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaled)
}
