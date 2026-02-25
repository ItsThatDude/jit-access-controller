package processors

import (
	"context"
	"errors"
	"fmt"
	"time"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	accessv1alpha1 "github.com/itsthatdude/jit-access-controller/api/v1alpha1"
	common "github.com/itsthatdude/jit-access-controller/internal/common"
	"github.com/itsthatdude/jit-access-controller/internal/metrics"
)

type GrantProcessor struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

func (r *GrantProcessor) ReconcileGrant(ctx context.Context, obj common.AccessGrantObject) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	objStatus := obj.GetStatus()

	originalStatus := *objStatus.DeepCopy()
	status := objStatus.DeepCopy()

	base := obj.DeepCopyObject().(client.Object)

	// Ensure status is persisted at the end of reconciliation
	defer func() {
		if obj.GetDeletionTimestamp().IsZero() {
			if !equality.Semantic.DeepEqual(originalStatus, *status) {
				obj.SetStatus(status)

				if err := r.Status().Patch(ctx, obj, client.MergeFrom(base)); err != nil {
					log.Error(err, "failed to persist status with patch")
				}
			}
		}
	}()

	// Handle deletion
	if !obj.GetDeletionTimestamp().IsZero() {
		log.Info("handling deletion of grant", "name", obj.GetName())

		err := r.handleExpired(ctx, obj, false)

		return ctrl.Result{}, err
	}

	// Add finalizer
	if obj.GetDeletionTimestamp().IsZero() {
		err := EnsureFinalizerExists(r.Client, ctx, obj, common.JITFinalizer)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				log.Info("grant not found when adding finalizer, it may have been deleted", "name", obj.GetName())
				return ctrl.Result{}, nil
			}

			log.Error(err, "an error occurred adding the finalizer to the grant", "name", obj.GetName())
			return ctrl.Result{}, err
		}
	}

	// If the RequestId is not set, the grant has only just been created.
	// Nothing to do until the status is populated.
	if status.RequestId == "" {
		return ctrl.Result{}, nil
	}

	// If the grant has expired, call handleExpired which cleans up the resources
	if !status.AccessExpiresAt.IsZero() && time.Now().After(status.AccessExpiresAt.Time) {
		err := r.handleExpired(ctx, obj, true)
		return ctrl.Result{}, err
	}

	return r.handleApproved(ctx, obj, status)
}

func (r *GrantProcessor) handleApproved(
	ctx context.Context,
	obj common.AccessGrantObject,
	status *accessv1alpha1.AccessGrantStatus,
) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	scope := obj.GetScope()

	// Handle pre-defined role or adhoc permissions
	isClusterScoped := scope == accessv1alpha1.RequestScopeCluster

	// Pre-defined Role/ClusterRole
	if status.Role.Name != "" && !status.RoleBindingCreated {
		roleBindingName := fmt.Sprintf("jit-access-%s", status.RequestId)

		if err := r.createRoleBinding(ctx, obj, status.Role, roleBindingName); err != nil && !k8serrors.IsAlreadyExists(err) {
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
			if err := r.createRole(ctx, obj, adhocName, status.Permissions); err != nil && !k8serrors.IsAlreadyExists(err) {
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

			if err := r.createRoleBinding(ctx, obj, rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: roleKind, Name: adhocName}, adhocName); err != nil && !k8serrors.IsAlreadyExists(err) {
				log.Error(err, "an error occurred creating the adhoc role binding for the request", "name", obj.GetName(), "subject", status.Subject, common.RoleKindRole, adhocName)
				return ctrl.Result{}, err
			}
			status.AdhocRoleBindingCreated = true
			log.Info("Created Adhoc Role Binding for request", "name", obj.GetName(), "subject", status.Subject, common.RoleKindRole, adhocName)
		}
	}

	// default duration fallback if not set
	durationStr := status.Duration
	if durationStr == "" {
		// nolint:goconst
		durationStr = "10m"
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		log.Error(err, "failed to parse duration string", "namespace", obj.GetNamespace(), "name", obj.GetName(), "duration", durationStr)
		return ctrl.Result{}, fmt.Errorf("failed to parse duration string: %w", err)
	}

	// Set expire time if not set
	if status.AccessExpiresAt.IsZero() {
		status.AccessExpiresAt = metav1.NewTime(time.Now().Add(duration))

		r.Recorder.Eventf(obj, "Normal", "AccessGranted",
			"Just-in-time access granted to %s for request %s",
			status.Subject, status.Request)
	}

	return ctrl.Result{
		RequeueAfter: time.Until(status.AccessExpiresAt.Time) + time.Second,
	}, nil
}

func (r *GrantProcessor) handleExpired(
	ctx context.Context,
	obj common.AccessGrantObject,
	deleteGrant bool,
) error {
	log := logf.FromContext(ctx)
	status := obj.GetStatus()

	// Clean up any resources created for this grant
	// Also cleans up the parent AccessRequest/ClusterAccessRequest object
	if err := r.cleanupResources(ctx, obj); err != nil {
		log.Error(err, "an error occurred running cleanup for the expired grant", "name", obj.GetName())
		return err
	}

	// Delete the grant object itself
	if deleteGrant && obj.GetDeletionTimestamp().IsZero() {
		log.Info("resources cleaned up for expired request, deleting the grant", "name", obj.GetName())
		if err := r.Delete(ctx, obj); err != nil && !k8serrors.IsNotFound(err) {
			log.Error(err, "failed to delete expired grant", "name", obj.GetName())
			return err
		}
	}

	// Remove the finalizer if it exists to allow deletion to complete
	if controllerutil.ContainsFinalizer(obj, common.JITFinalizer) {
		if err := RemoveFinalizer(r.Client, ctx, obj, common.JITFinalizer); err != nil && !k8serrors.IsNotFound(err) {
			log.Error(err, "an error occurred removing the grant finalizer", "name", obj.GetName())
			return err
		}
		log.Info("Removed finalizer for grant", "name", obj.GetName())
	}

	// Record an event about the revocation of access
	r.Recorder.Eventf(obj, "Normal", "AccessRevoked",
		"Just-in-time access revoked from %s for request %s",
		status.Subject, status.Request)

	metrics.GrantDuration.WithLabelValues(
		string(obj.GetScope()),
		obj.GetNamespace(),
		obj.GetName(),
		status.Subject,
	).Set(time.Since(obj.GetCreationTimestamp().Time).Seconds())

	return nil
}

func (r *GrantProcessor) cleanupResources(ctx context.Context, obj common.AccessGrantObject) error {
	log := logf.FromContext(ctx)
	status := obj.GetStatus()
	scope := obj.GetScope()
	requestId := status.RequestId

	log.Info("Cleaning up resources for grant", "name", obj.GetName(), "requestId", requestId)

	var errs []error

	deleteResource := func(key client.ObjectKey, obj client.Object, description string) {
		if err := r.Get(ctx, key, obj); err == nil {
			if err := r.Delete(ctx, obj); err != nil && !k8serrors.IsNotFound(err) {
				errs = append(errs, fmt.Errorf("failed to delete %s %s: %w", description, key.Name, err))
			} else {
				log.Info("Deleted "+description, "name", key.Name)
			}
		} else if !k8serrors.IsNotFound(err) {
			errs = append(errs, fmt.Errorf("failed to get %s %s: %w", description, key.Name, err))
		}
	}

	// Regular RoleBinding / ClusterRoleBinding
	if status.RoleBindingCreated {
		key := client.ObjectKey{Name: fmt.Sprintf("jit-access-%s", requestId)}
		var rb client.Object
		var desc string
		if scope == accessv1alpha1.RequestScopeCluster {
			rb = &rbacv1.ClusterRoleBinding{}
			desc = "ClusterRoleBinding"
		} else {
			rb = &rbacv1.RoleBinding{}
			key.Namespace = obj.GetNamespace()
			desc = "RoleBinding"
		}
		deleteResource(key, rb, desc)
	}

	// Adhoc RoleBinding / ClusterRoleBinding
	if status.AdhocRoleBindingCreated {
		key := client.ObjectKey{Name: fmt.Sprintf("jit-access-adhoc-%s", requestId)}
		var rb client.Object
		var desc string
		if scope == accessv1alpha1.RequestScopeCluster {
			rb = &rbacv1.ClusterRoleBinding{}
			desc = "ClusterRoleBinding"
		} else {
			rb = &rbacv1.RoleBinding{}
			key.Namespace = obj.GetNamespace()
			desc = "RoleBinding"
		}
		deleteResource(key, rb, fmt.Sprintf("Adhoc %s", desc))
	}

	// Adhoc Role / ClusterRole
	if status.AdhocRoleCreated {
		key := client.ObjectKey{Name: fmt.Sprintf("jit-access-adhoc-%s", requestId)}
		var roleObj client.Object
		var desc string
		if scope == accessv1alpha1.RequestScopeCluster {
			roleObj = &rbacv1.ClusterRole{}
			desc = common.RoleKindCluster
		} else {
			roleObj = &rbacv1.Role{}
			key.Namespace = obj.GetNamespace()
			desc = common.RoleKindRole
		}
		deleteResource(key, roleObj, fmt.Sprintf("Adhoc %s", desc))
	}

	// Delete the Request object
	reqKey := client.ObjectKey{Name: status.Request}
	var reqObj client.Object
	var reqType string
	if scope == accessv1alpha1.RequestScopeCluster {
		reqObj = &accessv1alpha1.ClusterAccessRequest{}
		reqType = "ClusterAccessRequest"
	} else {
		reqObj = &accessv1alpha1.AccessRequest{}
		reqKey.Namespace = obj.GetNamespace()
		reqType = "AccessRequest"
	}
	deleteResource(reqKey, reqObj, reqType)

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	log.Info("Resource cleanup complete for grant", "name", obj.GetName(), "requestId", requestId)

	return nil
}

func (r *GrantProcessor) createRole(
	ctx context.Context,
	obj common.AccessGrantObject,
	name string,
	rules []rbacv1.PolicyRule,
) error {
	scope := obj.GetScope()
	ns := obj.GetNamespace()
	labels := common.CommonLabels()

	isClusterScoped := scope == accessv1alpha1.RequestScopeCluster && ns == ""
	var role client.Object

	if isClusterScoped {
		role = &rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{Name: name, Labels: labels},
			Rules:      rules,
		}
	} else {
		role = &rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: obj.GetNamespace(), Labels: labels},
			Rules:      rules,
		}
	}

	if err := controllerutil.SetControllerReference(obj, role.(metav1.Object), r.Scheme); err != nil {
		return fmt.Errorf("failed to set owner reference on Role %s: %w", name, err)
	}

	return r.Create(ctx, role)
}

func (r *GrantProcessor) createRoleBinding(
	ctx context.Context,
	obj common.AccessGrantObject,
	roleRef rbacv1.RoleRef,
	bindingName string,
) error {
	scope := obj.GetScope()
	status := obj.GetStatus()
	labels := common.CommonLabels()

	isClusterScoped := scope == accessv1alpha1.RequestScopeCluster
	subject := rbacv1.Subject{Kind: "User", Name: status.Subject, APIGroup: "rbac.authorization.k8s.io"}

	var roleBinding client.Object

	if isClusterScoped {
		if roleRef.Kind != common.RoleKindCluster {
			return fmt.Errorf("can not bind a Role via ClusterRoleBinding")
		}
		roleBinding = &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: bindingName, Labels: labels},
			Subjects:   []rbacv1.Subject{subject},
			RoleRef:    rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: common.RoleKindCluster, Name: roleRef.Name},
		}
	} else {
		roleBinding = &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: bindingName, Namespace: obj.GetNamespace(), Labels: labels},
			Subjects:   []rbacv1.Subject{subject},
			RoleRef:    rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: roleRef.Kind, Name: roleRef.Name},
		}
	}

	if err := controllerutil.SetControllerReference(obj, roleBinding, r.Scheme); err != nil {
		return fmt.Errorf("failed to set owner reference on RoleBinding %s: %w", bindingName, err)
	}

	return r.Create(ctx, roleBinding)
}
