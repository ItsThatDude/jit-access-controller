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

package controller

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	accessv1alpha1 "github.com/itsthatdude/jit-access-controller/api/v1alpha1"
	"github.com/itsthatdude/jit-access-controller/internal/notifications"
	"github.com/nikoksr/notify"
)

// NotificationConfigReconciler reconciles a NotificationConfig object
type NotificationConfigReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Manager *notifications.NotificationManager
}

// +kubebuilder:rbac:groups=access.antware.xyz,resources=notificationconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=access.antware.xyz,resources=notificationconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=access.antware.xyz,resources=notificationconfigs/finalizers,verbs=update

func (r *NotificationConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	var cfg accessv1alpha1.NotificationConfig
	var name = req.Name
	if err := r.Get(ctx, req.NamespacedName, &cfg); err != nil {
		if apierrors.IsNotFound(err) {
			r.Manager.Update(name, nil)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	notifier, err := buildNotifier(ctx, r.Client, &cfg)
	if err != nil {
		return ctrl.Result{}, err
	}

	r.Manager.Update(name, notifier)
	return ctrl.Result{}, nil
}

func buildNotifier(
	ctx context.Context,
	c client.Client,
	cfg *accessv1alpha1.NotificationConfig,
) (notifications.Notifier, error) {
	n := notify.New()

	if cfg.Spec.Providers.Slack.Enabled {
		/*url, err := readSecret(
			ctx, c,
			cfg.Spec.Providers.Slack.WebhookURLSecretRef,
		)
		if err != nil {
			return nil, err
		}

		slackSvc := slack.New(url)
		n.UseServices(slackSvc)*/
	}

	// add more providers here

	return notifications.NewNotifier(n), nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NotificationConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&accessv1alpha1.NotificationConfig{}).
		Named("notificationconfig").
		Complete(r)
}
