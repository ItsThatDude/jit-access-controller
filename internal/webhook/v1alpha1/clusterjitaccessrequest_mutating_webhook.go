package v1alpha1

import (
	"context"
	"encoding/json"
	"net/http"

	"antware.xyz/jitaccess/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/mutate-access-antware-xyz-v1alpha1-clusterjitaccessrequest,mutating=true,failurePolicy=fail,sideEffects=None,groups=access.antware.xyz,resources=clusterjitaccessrequests,verbs=create;update,versions=v1alpha1,name=mclusterjitaccessrequest-v1alpha1.kb.io,admissionReviewVersions=v1

type ClusterJITAccessRequestMutator struct {
	decoder admission.Decoder
}

func SetupClusterJITAccessRequestMutatingWebhookWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(
		"/mutate-access-antware-xyz-v1alpha1-clusterjitaccessrequest",
		&admission.Webhook{Handler: &JITAccessRequestMutator{
			decoder: admission.NewDecoder(mgr.GetScheme()),
		}},
	)
}

func (m *ClusterJITAccessRequestMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &v1alpha1.ClusterJITAccessRequest{}

	if err := m.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Set the subject field if not already set
	if obj.Spec.Subject == "" {
		obj.Spec.Subject = req.UserInfo.Username
	}

	marshaled, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaled)
}
