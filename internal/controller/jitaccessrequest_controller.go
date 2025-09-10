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
	"fmt"
	"time"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	accessv1alpha1 "antware.xyz/jitaccess/api/v1alpha1"
	"antware.xyz/jitaccess/internal/policy"
	"antware.xyz/jitaccess/internal/utils"
)

// JITAccessRequestReconciler reconciles a JITAccessRequest object
type JITAccessRequestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const jitFinalizer = "access.antware.xyz/finalizer"

// +kubebuilder:rbac:groups=access.antware.xyz,resources=jitaccesspolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=access.antware.xyz,resources=jitaccesspolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=access.antware.xyz,resources=jitaccesspolicies/finalizers,verbs=update
// +kubebuilder:rbac:groups=access.antware.xyz,resources=jitaccessrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=access.antware.xyz,resources=jitaccessrequests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=access.antware.xyz,resources=jitaccessrequests/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the JITAccessRequest object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *JITAccessRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var jit accessv1alpha1.JITAccessRequest
	if err := r.Get(ctx, req.NamespacedName, &jit); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if jit.Status.RequestId == "" {
		jit.Status.RequestId = utils.GenerateRandomId()
		if err := r.Status().Update(ctx, &jit); err != nil {
			return ctrl.Result{}, err
		}
	}

	if jit.DeletionTimestamp.IsZero() && !controllerutil.ContainsFinalizer(&jit, jitFinalizer) {
		controllerutil.AddFinalizer(&jit, jitFinalizer)
		if err := r.Update(ctx, &jit); err != nil {
			return ctrl.Result{}, err
		}
		log.Info("Added finalizer to JITAccessRequest", "name", jit.Name)
	}

	// Handle deletion
	if !jit.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(&jit, jitFinalizer) {
			// Cleanup any role bindings left behind
			if err := r.cleanupResources(ctx, &jit); err != nil {
				return ctrl.Result{}, err
			}

			// Remove finalizer to allow deletion to complete
			controllerutil.RemoveFinalizer(&jit, jitFinalizer)
			if err := r.Update(ctx, &jit); err != nil {
				return ctrl.Result{}, err
			}

			log.Info("Cleaned up and removed finalizer", "name", jit.Name)
		}

		// Nothing more to do, let deletion continue
		return ctrl.Result{}, nil
	}

	// Check if the JIT Access Request matches a defined policy
	var policies accessv1alpha1.JITAccessPolicyList
	if err := r.List(ctx, &policies); err != nil {
		return ctrl.Result{}, err
	}

	permitted := policy.IsNamespacedRequestValid(&jit, &policies)

	if !permitted {
		log.Info("Access denied: no matching policy")
		// Optionally set a condition or delete the request
		return ctrl.Result{}, nil
	}

	if jit.Status.State == accessv1alpha1.RequestStateApproved {
		if jit.Status.ExpiresAt != nil {
			// If already granted and expired, clean up
			if time.Now().After(jit.Status.ExpiresAt.Time) {
				if err := r.cleanupResources(ctx, &jit); err != nil {
					return ctrl.Result{}, err
				}

				_ = r.Delete(ctx, &jit)

				log.Info("Access expired and cleaned up", "name", req.Name)
				return ctrl.Result{}, nil
			}
			// Not expired yet, requeue until then
			return ctrl.Result{RequeueAfter: time.Until(jit.Status.ExpiresAt.Time)}, nil
		} else {
			// Reconcile the approved access request
			roleBinding := &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("jit-access-%s", jit.Status.RequestId),
					Namespace: req.Namespace,
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:     "User",
						Name:     jit.Spec.Subject,
						APIGroup: "rbac.authorization.k8s.io",
					},
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "Role",
					Name:     jit.Spec.Role,
				},
			}

			if err := r.Create(ctx, roleBinding); err != nil && !errors.IsAlreadyExists(err) {
				log.Error(err, "failed to create RoleBinding")
				return ctrl.Result{}, err
			}

			// Update status
			expireTime := metav1.NewTime(time.Now().Add(time.Duration(jit.Spec.DurationSeconds) * time.Second))
			jit.Status.ExpiresAt = &expireTime
			if err := r.Status().Update(ctx, &jit); err != nil {
				log.Error(err, "failed to update status")
				return ctrl.Result{}, err
			}

			log.Info("Granted access", "subject", jit.Spec.Subject, "role", jit.Spec.Role)

			return ctrl.Result{RequeueAfter: time.Duration(jit.Spec.DurationSeconds) * time.Second}, nil
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *JITAccessRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&accessv1alpha1.JITAccessRequest{}).
		Named("jitaccessrequest").
		Complete(r)
}

func (r *JITAccessRequestReconciler) cleanupResources(ctx context.Context, jit *accessv1alpha1.JITAccessRequest) error {
	log := logf.FromContext(ctx)

	rbName := fmt.Sprintf("jit-access-%s", jit.Status.RequestId)
	rb := &rbacv1.RoleBinding{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      rbName,
		Namespace: jit.Namespace,
	}, rb)

	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if err == nil {
		// RoleBinding exists, delete it
		if err := r.Delete(ctx, rb); err != nil {
			return err
		}
		log.Info("Deleted RoleBinding for JITAccessRequest", "rolebinding", rbName)
	}

	return nil
}
