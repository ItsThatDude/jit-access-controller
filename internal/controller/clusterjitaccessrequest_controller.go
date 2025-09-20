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

	accessv1alpha1 "antware.xyz/jitaccess/api/v1alpha1"
	common "antware.xyz/jitaccess/internal/common"
	"antware.xyz/jitaccess/internal/policy"
	"antware.xyz/jitaccess/internal/utils"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	set "k8s.io/apimachinery/pkg/util/sets"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ClusterJITAccessRequestReconciler reconciles a ClusterJITAccessRequest object
type ClusterJITAccessRequestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusterjitaccesspolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusterjitaccesspolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusterjitaccesspolicies/finalizers,verbs=update
// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusterjitaccessrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusterjitaccessrequests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusterjitaccessrequests/finalizers,verbs=update
// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusterjitaccessresponses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusterjitaccessresponses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusterjitaccessresponses/finalizers,verbs=update
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;delete;bind;escalate
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create;delete;bind;escalate

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ClusterJITAccessRequest object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *ClusterJITAccessRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var jit accessv1alpha1.ClusterJITAccessRequest
	if err := r.Get(ctx, req.NamespacedName, &jit); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// If the requestId is not set, the request is new
	if jit.Status.RequestId == "" {
		jit.Status.RequestId = utils.GenerateRandomId()
		jit.Status.State = accessv1alpha1.RequestStatePending
		if err := r.Status().Update(ctx, &jit); err != nil {
			return ctrl.Result{}, err
		}
	}

	if jit.DeletionTimestamp.IsZero() && !controllerutil.ContainsFinalizer(&jit, common.JITFinalizer) {
		patch := client.MergeFrom(jit.DeepCopy())

		controllerutil.AddFinalizer(&jit, common.JITFinalizer)

		if err := r.Patch(ctx, &jit, patch); err != nil {
			return ctrl.Result{}, err
		}

		log.Info("Added finalizer to JITAccessRequest", "name", jit.Name)
	}

	// Handle deletion
	if !jit.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(&jit, common.JITFinalizer) {
			// Cleanup any role bindings left behind
			if err := r.cleanupResources(ctx, &jit); err != nil {
				return ctrl.Result{}, err
			}

			// Remove finalizer to allow deletion to complete
			controllerutil.RemoveFinalizer(&jit, common.JITFinalizer)
			if err := r.Update(ctx, &jit); err != nil {
				return ctrl.Result{}, err
			}

			log.Info("Cleaned up and removed finalizer", "name", jit.Name)
		}

		// Nothing more to do, let deletion continue
		return ctrl.Result{}, nil
	}

	// Check if the JIT Access Request matches a defined policy
	var policies accessv1alpha1.ClusterJITAccessPolicyList
	if err := r.List(ctx, &policies); err != nil {
		return ctrl.Result{}, err
	}

	isRequestValid, matched_policy := policy.IsClusterRequestValid(&jit, &policies)

	if !isRequestValid {
		log.Info("Access denied: no matching policy")
		// Optionally set a condition or delete the request
		return ctrl.Result{}, nil
	}

	if matched_policy.RequiredApprovals != jit.Status.ApprovalsRequired {
		jit.Status.ApprovalsRequired = matched_policy.RequiredApprovals
		if err := r.Status().Update(ctx, &jit); err != nil {
			log.Error(err, "failed to update status")
			return ctrl.Result{Requeue: true}, err
		}
	}

	switch jit.Status.State {
	case accessv1alpha1.RequestStateApproved:
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

			// Update status
			expireTime := metav1.NewTime(time.Now().Add(time.Duration(jit.Spec.DurationSeconds) * time.Second))
			jit.Status.ExpiresAt = &expireTime

			if err := r.Status().Update(ctx, &jit); err != nil {
				log.Error(err, "failed to update status")
				return ctrl.Result{}, err
			}

			if jit.Spec.Role != "" {
				roleBinding := &rbacv1.ClusterRoleBinding{
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
						Kind:     "ClusterRole",
						Name:     jit.Spec.Role,
					},
				}

				if err := r.Create(ctx, roleBinding); err != nil && !errors.IsAlreadyExists(err) {
					log.Error(err, "failed to create ClusterRoleBinding")
					return ctrl.Result{}, err
				}

				jit.Status.RoleBindingCreated = true

				if err := r.Status().Update(ctx, &jit); err != nil {
					log.Error(err, "failed to update status")
					return ctrl.Result{}, err
				}

				log.Info("Granted access", "subject", jit.Spec.Subject, "role", jit.Spec.Role)
			}

			if len(jit.Spec.Permissions) > 0 {
				name := fmt.Sprintf("jit-access-adhoc-%s", jit.Status.RequestId)

				role := &rbacv1.Role{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: req.Namespace,
					},
					Rules: jit.Spec.Permissions,
				}

				if err := r.Create(ctx, role); err != nil && !errors.IsAlreadyExists(err) {
					log.Error(err, "failed to create ClusterRole")
					return ctrl.Result{}, err
				}

				log.Info("ClusterRole created", "namespace", req.Namespace, "role", name)

				jit.Status.AdhocRoleCreated = true

				if err := r.Status().Update(ctx, &jit); err != nil {
					log.Error(err, "failed to update status")
					return ctrl.Result{}, err
				}

				roleBinding := &rbacv1.ClusterRoleBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name: name,
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
						Kind:     "ClusterRole",
						Name:     name,
					},
				}

				if err := r.Create(ctx, roleBinding); err != nil && !errors.IsAlreadyExists(err) {
					log.Error(err, "failed to create ClusterRoleBinding")
					return ctrl.Result{}, err
				}

				log.Info("ClusterRoleBinding Created", "namespace", req.Namespace, "role", name)

				jit.Status.AdhocRoleBindingCreated = true

				if err := r.Status().Update(ctx, &jit); err != nil {
					log.Error(err, "failed to update status")
					return ctrl.Result{}, err
				}
			}

			return ctrl.Result{RequeueAfter: time.Duration(jit.Spec.DurationSeconds) * time.Second}, nil
		}
	case accessv1alpha1.RequestStatePending:
		responses := &accessv1alpha1.ClusterJITAccessResponseList{}
		if err := r.List(ctx, responses, client.MatchingFields{"spec.requestRef": jit.Name}); err != nil {
			return ctrl.Result{}, err
		}

		approved := set.New[string]()
		denied := set.New[string]()

		for _, r := range responses.Items {
			if r.Spec.Response == accessv1alpha1.ResponseStateApproved {
				approved.Insert(r.Spec.Approver)
			}
			if r.Spec.Response == accessv1alpha1.ResponseStateDenied {
				denied.Insert(r.Spec.Approver)
			}
		}

		updateStatus := false

		if approved.Len() != jit.Status.ApprovalsReceived {
			jit.Status.ApprovalsReceived = approved.Len()
			updateStatus = true
		}

		if denied.Len() > 0 {
			jit.Status.State = accessv1alpha1.RequestStateDenied
			updateStatus = true
		} else if approved.Len() >= matched_policy.RequiredApprovals {
			jit.Status.State = accessv1alpha1.RequestStateApproved
			updateStatus = true
		}

		if updateStatus {
			if err := r.Status().Update(ctx, &jit); err != nil {
				log.Error(err, "failed to update status")
				return ctrl.Result{Requeue: true}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterJITAccessRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()
	indexer := mgr.GetFieldIndexer()

	if err := indexer.IndexField(ctx, &accessv1alpha1.ClusterJITAccessRequest{}, "status.requestId",
		func(obj client.Object) []string {
			if myObj, ok := obj.(*accessv1alpha1.ClusterJITAccessRequest); ok {
				return []string{myObj.Status.RequestId}
			}
			return nil
		}); err != nil {
		return fmt.Errorf("failed to add index for requestId: %w", err)
	}

	if err := indexer.IndexField(ctx, &accessv1alpha1.ClusterJITAccessResponse{}, "spec.requestRef",
		func(obj client.Object) []string {
			if myObj, ok := obj.(*accessv1alpha1.ClusterJITAccessResponse); ok {
				return []string{myObj.Spec.RequestRef}
			}
			return nil
		}); err != nil {
		return fmt.Errorf("failed to add index for requestRef: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&accessv1alpha1.ClusterJITAccessRequest{}).
		Watches(
			&accessv1alpha1.ClusterJITAccessResponse{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				resp := obj.(*accessv1alpha1.ClusterJITAccessResponse)
				return []reconcile.Request{{
					NamespacedName: types.NamespacedName{
						Name: resp.Spec.RequestRef,
					},
				}}
			}),
		).
		Named("clusterjitaccessrequest").
		Complete(r)
}

func (r *ClusterJITAccessRequestReconciler) cleanupResources(ctx context.Context, jit *accessv1alpha1.ClusterJITAccessRequest) error {
	log := logf.FromContext(ctx)

	// If a regular Role Binding was created, delete it
	if jit.Status.RoleBindingCreated {
		rbName := fmt.Sprintf("jit-access-%s", jit.Status.RequestId)
		rb := &rbacv1.ClusterRoleBinding{}
		err := r.Get(ctx, client.ObjectKey{Name: rbName}, rb)

		if err != nil && !errors.IsNotFound(err) {
			return err
		}

		if err == nil {
			// RoleBinding exists, delete it
			if err := r.Delete(ctx, rb); err != nil {
				return err
			}
			log.Info("Deleted RoleBinding for JITAccessRequest", "clusterrolebinding", rbName)
		}
	}

	// If an Adhoc Role Binding was created, delete it
	if jit.Status.AdhocRoleBindingCreated {
		rbName := fmt.Sprintf("jit-access-adhoc-%s", jit.Status.RequestId)
		rb := &rbacv1.ClusterRoleBinding{}
		err := r.Get(ctx, client.ObjectKey{Name: rbName}, rb)

		if err != nil && !errors.IsNotFound(err) {
			return err
		}

		if err == nil {
			// RoleBinding exists, delete it
			if err := r.Delete(ctx, rb); err != nil {
				return err
			}
			log.Info("Deleted Adhoc RoleBinding for JITAccessRequest", "clusterrolebinding", rbName)
		}
	}

	// If an Adhoc Role was created, delete it
	if jit.Status.AdhocRoleCreated {
		roleName := fmt.Sprintf("jit-access-adhoc-%s", jit.Status.RequestId)
		role := &rbacv1.ClusterRole{}
		err := r.Get(ctx, client.ObjectKey{Name: roleName}, role)

		if err != nil && !errors.IsNotFound(err) {
			return err
		}

		if err == nil {
			// Role exists, delete it
			if err := r.Delete(ctx, role); err != nil {
				return err
			}
			log.Info("Deleted Adhoc Role for ClusterJITAccessRequest", "clusterrole", roleName)
		}
	}

	// Delete all of the responses for this request
	responses := &accessv1alpha1.ClusterJITAccessResponseList{}
	if err := r.List(ctx, responses,
		client.MatchingFields{"spec.requestRef": jit.Name}); err != nil {
		return err
	}

	for _, resp := range responses.Items {
		log.Info("Deleting ClusterJITAccessResponse", "ClusterJITAccessResponse", resp.Name)
		if err := r.Delete(ctx, &resp); err != nil {
			return err
		}
	}

	return nil
}
