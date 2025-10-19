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
	goerr "errors"
	"fmt"
	"time"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	accessv1alpha1 "antware.xyz/jitaccess/api/v1alpha1"
	common "antware.xyz/jitaccess/internal/common"
)

// GrantReconciler reconciles a JITAccessGrant object
type GrantReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	Recorder        record.EventRecorder
	SystemNamespace string
}

// +kubebuilder:rbac:groups=access.antware.xyz,resources=jitaccessgrants,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=access.antware.xyz,resources=jitaccessgrants/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=access.antware.xyz,resources=jitaccessgrants/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the JITAccessGrant object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *GrantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var grant accessv1alpha1.JITAccessGrant
	err := r.Get(ctx, req.NamespacedName, &grant)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	log := logf.FromContext(ctx)

	originalStatus := *grant.Status.DeepCopy()
	status := grant.Status.DeepCopy()

	base := grant.DeepCopyObject().(client.Object)

	defer func() {
		if grant.GetDeletionTimestamp().IsZero() {
			if !equality.Semantic.DeepEqual(originalStatus, *status) {
				grant.Status = *status

				if err := r.Status().Patch(ctx, &grant, client.MergeFrom(base)); err != nil {
					log.Error(err, "failed to persist status with patch")
				}
			}
		}
	}()

	// Add finalizer
	if grant.GetDeletionTimestamp().IsZero() && !controllerutil.ContainsFinalizer(&grant, common.JITFinalizer) {
		if err := r.ensureFinalizer(ctx, &grant, common.JITFinalizer); err != nil {
			log.Error(err, "an error occurred updating the finalizer for the request", "name", grant.GetName())
			return ctrl.Result{}, err
		}
		log.Info("Added finalizer to request", "name", grant.GetName())
	}

	// Handle deletion
	if !grant.GetDeletionTimestamp().IsZero() {
		if controllerutil.ContainsFinalizer(&grant, common.JITFinalizer) {
			if err := r.cleanupResources(ctx, &grant); err != nil {
				log.Error(err, "an error occurred running cleanup for the request", "name", grant.GetName())
				return ctrl.Result{}, err
			}
			if err := r.removeFinalizer(ctx, &grant, common.JITFinalizer); err != nil {
				log.Error(err, "an error occurred removing the request finalizer", "name", grant.GetName())
				return ctrl.Result{}, err
			}
			log.Info("Cleaned up and removed finalizer", "name", grant.GetName())
		}
		return ctrl.Result{}, nil
	}

	if status.RequestId == "" {
		return ctrl.Result{}, nil
	}

	if status.AccessExpiresAt != nil && time.Now().After(status.AccessExpiresAt.Time) {
		return r.handleExpired(ctx, &grant)
	}

	return r.handleApproved(ctx, &grant, status)
}

func (r *GrantReconciler) ensureFinalizer(ctx context.Context, obj client.Object, finalizer string) error {
	if obj.GetDeletionTimestamp().IsZero() && !controllerutil.ContainsFinalizer(obj, finalizer) {
		patch := client.MergeFrom(obj.DeepCopyObject().(client.Object))
		controllerutil.AddFinalizer(obj, finalizer)
		if err := r.Patch(ctx, obj, patch); err != nil {
			return fmt.Errorf("failed to add finalizer: %w", err)
		}
	}
	return nil
}

func (r *GrantReconciler) removeFinalizer(ctx context.Context, obj client.Object, finalizer string) error {
	if controllerutil.ContainsFinalizer(obj, finalizer) {
		patch := client.MergeFrom(obj.DeepCopyObject().(client.Object))
		controllerutil.RemoveFinalizer(obj, finalizer)
		return r.Patch(ctx, obj, patch)
	}
	return nil
}

func (r *GrantReconciler) handleApproved(
	ctx context.Context,
	obj *accessv1alpha1.JITAccessGrant,
	status *accessv1alpha1.JITAccessGrantStatus,
) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Handle pre-defined role or adhoc permissions
	isClusterScoped := obj.Status.Namespace == ""

	// Pre-defined Role/ClusterRole
	if status.Role.Name != "" && !status.RoleBindingCreated {
		roleBindingName := fmt.Sprintf("jit-access-%s", status.RequestId)

		if err := r.createRoleBinding(ctx, obj, status.Role, roleBindingName, isClusterScoped); err != nil && !errors.IsAlreadyExists(err) {
			log.Error(err, "an error occurred creating the role binding for the request", "name", obj.GetName(), "subject", status.Subject, common.RoleKindRole, status.Role)
			return ctrl.Result{}, err
		}
		status.RoleBindingCreated = true
		log.Info("Granted Role for request", "name", obj.GetName(), "subject", status.Subject, common.RoleKindRole, status.Role)
	}

	// Adhoc permissions
	if len(status.Permissions) > 0 && (!status.AdhocRoleCreated || !status.AdhocRoleBindingCreated) {
		adhocName := fmt.Sprintf("jit-access-adhoc-%s", status.RequestId)

		if !status.AdhocRoleCreated {
			if err := r.createRole(ctx, obj, adhocName, status.Permissions, isClusterScoped); err != nil && !errors.IsAlreadyExists(err) {
				log.Error(err, "an error occurred creating the adhoc role for the request", "name", obj.GetName(), "subject", status.Subject, common.RoleKindRole, adhocName)
				return ctrl.Result{}, err
			}
			status.AdhocRoleCreated = true
			log.Info("Created Adhoc Role for request", "name", obj.GetName(), "subject", status.Subject, common.RoleKindRole, adhocName)
		}

		if !status.AdhocRoleBindingCreated {
			roleKind := common.RoleKindRole

			if isClusterScoped {
				roleKind = common.RoleKindCluster
			}

			if err := r.createRoleBinding(ctx, obj, rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: roleKind, Name: adhocName}, adhocName, isClusterScoped); err != nil && !errors.IsAlreadyExists(err) {
				log.Error(err, "an error occurred creating the adhoc role binding for the request", "name", obj.GetName(), "subject", status.Subject, common.RoleKindRole, adhocName)
				return ctrl.Result{}, err
			}
			status.AdhocRoleBindingCreated = true
			log.Info("Created Adhoc Role Binding for request", "name", obj.GetName(), "subject", status.Subject, common.RoleKindRole, adhocName)
		}
	}

	// Set expire time if not set
	if status.AccessExpiresAt == nil {
		expireTime := metav1.NewTime(time.Now().Add(time.Duration(status.DurationSeconds) * time.Second))
		status.AccessExpiresAt = &expireTime
	}

	r.Recorder.Eventf(obj, "Normal", "AccessGranted",
		"Just-in-time access granted to %s for request %s",
		obj.Status.Subject, obj.Status.Request)

	return ctrl.Result{RequeueAfter: time.Duration(status.DurationSeconds) * time.Second}, nil
}

func (r *GrantReconciler) handleExpired(
	ctx context.Context,
	obj *accessv1alpha1.JITAccessGrant,
) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	if err := r.cleanupResources(ctx, obj); err != nil {
		log.Error(err, "an error occurred running cleanup for the expired request", "name", obj.GetName())
		return ctrl.Result{RequeueAfter: 10}, err
	}

	log.Info("resources cleaned up for expired request, deleting the request", "name", obj.GetName())
	_ = r.Delete(ctx, obj)

	return ctrl.Result{}, nil
}

func (r *GrantReconciler) cleanupResources(ctx context.Context, obj *accessv1alpha1.JITAccessGrant) error {
	log := logf.FromContext(ctx)
	requestId := obj.Status.RequestId

	var errs []error

	deleteResource := func(key client.ObjectKey, obj client.Object, description string) {
		if err := r.Get(ctx, key, obj); err == nil {
			if err := r.Delete(ctx, obj); err != nil {
				errs = append(errs, fmt.Errorf("failed to delete %s %s: %w", description, key.Name, err))
			} else {
				log.Info("Deleted "+description, "name", key.Name)
			}
		} else if !errors.IsNotFound(err) {
			errs = append(errs, fmt.Errorf("failed to get %s %s: %w", description, key.Name, err))
		}
	}

	// Regular RoleBinding / ClusterRoleBinding
	if obj.Status.RoleBindingCreated {
		key := client.ObjectKey{Name: fmt.Sprintf("jit-access-%s", requestId)}
		var rb client.Object
		var desc string
		if obj.Status.Scope == accessv1alpha1.GrantScopeCluster {
			rb = &rbacv1.ClusterRoleBinding{}
			desc = "ClusterRoleBinding"
		} else {
			rb = &rbacv1.RoleBinding{}
			key.Namespace = obj.Status.Namespace
			desc = "RoleBinding"
		}
		deleteResource(key, rb, desc)
	}

	// Adhoc RoleBinding / ClusterRoleBinding
	if obj.Status.AdhocRoleBindingCreated {
		key := client.ObjectKey{Name: fmt.Sprintf("jit-access-adhoc-%s", requestId)}
		var rb client.Object
		var desc string
		if obj.Status.Scope == accessv1alpha1.GrantScopeCluster {
			rb = &rbacv1.ClusterRoleBinding{}
			desc = "ClusterRoleBinding"
		} else {
			rb = &rbacv1.RoleBinding{}
			key.Namespace = obj.Status.Namespace
			desc = "RoleBinding"
		}
		deleteResource(key, rb, fmt.Sprintf("Adhoc %s", desc))
	}

	// Adhoc Role / ClusterRole
	if obj.Status.AdhocRoleCreated {
		key := client.ObjectKey{Name: fmt.Sprintf("jit-access-adhoc-%s", requestId)}
		var roleObj client.Object
		var desc string
		if obj.Status.Scope == accessv1alpha1.GrantScopeCluster {
			roleObj = &rbacv1.ClusterRole{}
			desc = common.RoleKindCluster
		} else {
			roleObj = &rbacv1.Role{}
			key.Namespace = obj.Status.Namespace
			desc = common.RoleKindRole
		}
		deleteResource(key, roleObj, fmt.Sprintf("Adhoc %s", desc))
	}

	// Delete the Request object
	reqKey := client.ObjectKey{Name: obj.Status.Request}
	var reqObj client.Object
	var reqType string
	if obj.Status.Scope == accessv1alpha1.GrantScopeCluster {
		reqObj = &accessv1alpha1.ClusterJITAccessRequest{}
		reqType = "ClusterJITAccessRequest"
	} else {
		reqObj = &accessv1alpha1.JITAccessRequest{}
		reqKey.Namespace = obj.Status.Namespace
		reqType = "JITAccessRequest"
	}
	deleteResource(reqKey, reqObj, reqType)

	if len(errs) > 0 {
		return goerr.Join(errs...)
	}

	r.Recorder.Eventf(obj, "Normal", "AccessRevoked",
		"Just-in-time access revoked from %s for request %s",
		obj.Status.Subject, obj.Status.Request)

	return nil
}

func (r *GrantReconciler) createRole(
	ctx context.Context,
	obj *accessv1alpha1.JITAccessGrant,
	name string,
	rules []rbacv1.PolicyRule,
	clusterRole bool,
) error {
	labels := common.CommonLabels()
	var role client.Object
	if clusterRole {
		role = &rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{Name: name, Labels: labels},
			Rules:      rules,
		}
	} else {
		role = &rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: obj.Status.Namespace, Labels: labels},
			Rules:      rules,
		}
	}
	return r.Create(ctx, role)
}

func (r *GrantReconciler) createRoleBinding(
	ctx context.Context,
	obj *accessv1alpha1.JITAccessGrant,
	roleRef rbacv1.RoleRef,
	bindingName string,
	clusterScoped bool,
) error {
	labels := common.CommonLabels()
	subject := rbacv1.Subject{Kind: "User", Name: obj.Status.Subject, APIGroup: "rbac.authorization.k8s.io"}
	if clusterScoped {
		if roleRef.Kind != common.RoleKindCluster {
			return fmt.Errorf("can not bind a Role via ClusterRoleBinding")
		}
		rb := &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: bindingName, Labels: labels},
			Subjects:   []rbacv1.Subject{subject},
			RoleRef:    rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: common.RoleKindCluster, Name: roleRef.Name},
		}
		return r.Create(ctx, rb)
	} else {
		rb := &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: bindingName, Namespace: obj.Status.Namespace, Labels: labels},
			Subjects:   []rbacv1.Subject{subject},
			RoleRef:    rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: roleRef.Kind, Name: roleRef.Name},
		}
		return r.Create(ctx, rb)
	}
}

func (r *GrantReconciler) SetupWithManagerNamespaced(mgr ctrl.Manager) error {
	ctx := context.Background()
	indexer := mgr.GetFieldIndexer()

	if err := indexer.IndexField(ctx, &accessv1alpha1.JITAccessGrant{}, "status.requestId",
		func(obj client.Object) []string {
			if myObj, ok := obj.(*accessv1alpha1.JITAccessGrant); ok {
				return []string{myObj.Status.RequestId}
			}
			return nil
		}); err != nil {
		return fmt.Errorf("failed to add index for requestId: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&accessv1alpha1.JITAccessGrant{}).
		Named("grant-reconciler-namespaced").
		Complete(r)
}
