package v1alpha1

import (
	"context"
	"encoding/json"
	"net/http"

	"antware.xyz/kairos/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/mutate-access-antware-xyz-v1alpha1-accessrequest,mutating=true,failurePolicy=fail,sideEffects=None,groups=access.antware.xyz,resources=accessrequests,verbs=create;update,versions=v1alpha1,name=maccessrequest-v1alpha1.kb.io,admissionReviewVersions=v1

type AccessRequestMutator struct {
	decoder admission.Decoder
}

func SetupAccessRequestMutatingWebhookWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(
		"/mutate-access-antware-xyz-v1alpha1-accessrequest",
		&admission.Webhook{Handler: &AccessRequestMutator{
			decoder: admission.NewDecoder(mgr.GetScheme()),
		}},
	)
}

func (m *AccessRequestMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &v1alpha1.AccessRequest{}

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
