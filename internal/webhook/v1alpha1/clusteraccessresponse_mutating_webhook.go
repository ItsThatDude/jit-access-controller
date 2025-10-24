package v1alpha1

import (
	"context"
	"encoding/json"
	"net/http"

	"antware.xyz/jitaccess/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/mutate-access-antware-xyz-v1alpha1-clusteraccessresponse,mutating=true,failurePolicy=fail,sideEffects=None,groups=access.antware.xyz,resources=clusteraccessresponses,verbs=create;update,versions=v1alpha1,name=mclusteraccessresponse-v1alpha1.kb.io,admissionReviewVersions=v1

type ClusterAccessResponseMutator struct {
	decoder admission.Decoder
}

func SetupClusterAccessResponseMutatingWebhookWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(
		"/mutate-access-antware-xyz-v1alpha1-clusteraccessresponse",
		&admission.Webhook{Handler: &ClusterAccessResponseMutator{
			decoder: admission.NewDecoder(mgr.GetScheme()),
		}},
	)
}

func (m *ClusterAccessResponseMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &v1alpha1.ClusterAccessResponse{}

	if err := m.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Set the approver to the current user
	obj.Spec.Approver = req.UserInfo.Username

	marshaled, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaled)
}
