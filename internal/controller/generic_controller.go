package controller

import (
	"context"
	goerr "errors"
	"fmt"
	"time"

	"antware.xyz/jitaccess/api/v1alpha1"
	common "antware.xyz/jitaccess/internal/common"
	"antware.xyz/jitaccess/internal/policy"
	"antware.xyz/jitaccess/internal/utils"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	set "k8s.io/apimachinery/pkg/util/sets"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusterjitaccesspolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusterjitaccesspolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusterjitaccesspolicies/finalizers,verbs=update
// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusterjitaccessrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusterjitaccessrequests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusterjitaccessrequests/finalizers,verbs=update
// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusterjitaccessresponses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusterjitaccessresponses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusterjitaccessresponses/finalizers,verbs=update

// +kubebuilder:rbac:groups=access.antware.xyz,resources=jitaccesspolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=access.antware.xyz,resources=jitaccesspolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=access.antware.xyz,resources=jitaccesspolicies/finalizers,verbs=update
// +kubebuilder:rbac:groups=access.antware.xyz,resources=jitaccessrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=access.antware.xyz,resources=jitaccessrequests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=access.antware.xyz,resources=jitaccessrequests/finalizers,verbs=update
// +kubebuilder:rbac:groups=access.antware.xyz,resources=jitaccessresponses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=access.antware.xyz,resources=jitaccessresponses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=access.antware.xyz,resources=jitaccessresponses/finalizers,verbs=update

// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;delete;bind;escalate
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create;delete;bind;escalate

type GenericJITAccessReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *GenericJITAccessReconciler) SetupWithManagerCluster(mgr ctrl.Manager) error {
	ctx := context.Background()
	indexer := mgr.GetFieldIndexer()

	if err := indexer.IndexField(ctx, &v1alpha1.ClusterJITAccessRequest{}, "status.requestId",
		func(obj client.Object) []string {
			if myObj, ok := obj.(*v1alpha1.ClusterJITAccessRequest); ok {
				return []string{myObj.Status.RequestId}
			}
			return nil
		}); err != nil {
		return fmt.Errorf("failed to add index for requestId: %w", err)
	}

	if err := indexer.IndexField(ctx, &v1alpha1.ClusterJITAccessResponse{}, "spec.requestRef",
		func(obj client.Object) []string {
			if myObj, ok := obj.(*v1alpha1.ClusterJITAccessResponse); ok {
				return []string{myObj.Spec.RequestRef}
			}
			return nil
		}); err != nil {
		return fmt.Errorf("failed to add index for requestRef: %w", err)
	}

	eventFilter := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ClusterJITAccessRequest{}).
		Watches(
			&v1alpha1.ClusterJITAccessResponse{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				resp := obj.(*v1alpha1.ClusterJITAccessResponse)
				return []reconcile.Request{{
					NamespacedName: types.NamespacedName{
						Name: resp.Spec.RequestRef,
					},
				}}
			}),
			builder.WithPredicates(eventFilter),
		).
		Named("jitreconciler-cluster").
		Complete(r)
}

func (r *GenericJITAccessReconciler) SetupWithManagerNamespaced(mgr ctrl.Manager) error {
	ctx := context.Background()
	indexer := mgr.GetFieldIndexer()

	if err := indexer.IndexField(ctx, &v1alpha1.JITAccessRequest{}, "status.requestId",
		func(obj client.Object) []string {
			if myObj, ok := obj.(*v1alpha1.JITAccessRequest); ok {
				return []string{myObj.Status.RequestId}
			}
			return nil
		}); err != nil {
		return fmt.Errorf("failed to add index for requestId: %w", err)
	}

	if err := indexer.IndexField(ctx, &v1alpha1.JITAccessResponse{}, "spec.requestRef",
		func(obj client.Object) []string {
			if myObj, ok := obj.(*v1alpha1.JITAccessResponse); ok {
				return []string{myObj.Spec.RequestRef}
			}
			return nil
		}); err != nil {
		return fmt.Errorf("failed to add index for requestRef: %w", err)
	}

	eventFilter := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.JITAccessRequest{}).
		Watches(
			&v1alpha1.JITAccessResponse{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				resp := obj.(*v1alpha1.JITAccessResponse)
				return []reconcile.Request{{
					NamespacedName: types.NamespacedName{
						Namespace: resp.Namespace,
						Name:      resp.Spec.RequestRef,
					},
				}}
			}),
			builder.WithPredicates(eventFilter),
		).
		Named("jitreconciler-namespaced").
		Complete(r)
}

func (r *GenericJITAccessReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if req.Namespace == "" {
		var clusterObj v1alpha1.ClusterJITAccessRequest
		err := r.Get(ctx, types.NamespacedName{Name: req.Name}, &clusterObj)
		if err != nil {
			if errors.IsNotFound(err) {
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, err
		}
		return r.reconcileGeneric(ctx, &clusterObj)
	}

	var nsObj v1alpha1.JITAccessRequest
	err := r.Get(ctx, req.NamespacedName, &nsObj)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	return r.reconcileGeneric(ctx, &nsObj)
}

func (r *GenericJITAccessReconciler) reconcileGeneric(ctx context.Context, obj common.JITAccessRequestObject) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	originalStatus := *obj.GetStatus().DeepCopy()
	status := obj.GetStatus().DeepCopy()

	base := obj.DeepCopyObject().(client.Object)

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

	// If RequestId not set, then it's a new request
	if status.RequestId == "" {
		status.RequestId = utils.GenerateRandomId()
		status.State = v1alpha1.RequestStatePending
	}

	// Add finalizer
	if obj.GetDeletionTimestamp().IsZero() && !controllerutil.ContainsFinalizer(obj, common.JITFinalizer) {
		if err := r.ensureFinalizer(ctx, obj, common.JITFinalizer); err != nil {
			log.Error(err, "an error occurred updating the finalizer for the request", "name", obj.GetName())
			return ctrl.Result{}, err
		}
		log.Info("Added finalizer to request", "name", obj.GetName())
	}

	// Handle deletion
	if !obj.GetDeletionTimestamp().IsZero() {
		if controllerutil.ContainsFinalizer(obj, common.JITFinalizer) {
			if err := r.cleanupResources(ctx, obj); err != nil {
				log.Error(err, "an error occurred running cleanup for the request", "name", obj.GetName())
				return ctrl.Result{}, err
			}
			if err := r.removeFinalizer(ctx, obj, common.JITFinalizer); err != nil {
				log.Error(err, "an error occurred removing the request finalizer", "name", obj.GetName())
				return ctrl.Result{}, err
			}
			log.Info("Cleaned up and removed finalizer", "name", obj.GetName())
		}
		return ctrl.Result{}, nil
	}

	// Match against policies
	ns := obj.GetNamespace()
	isValid := false
	var matched *v1alpha1.SubjectPolicy
	if ns == "" {
		var clusterPolicies v1alpha1.ClusterJITAccessPolicyList
		if err := r.List(ctx, &clusterPolicies); err != nil {
			return ctrl.Result{}, err
		}

		isValid, matched = policy.IsRequestValid(obj, clusterPolicies.Items)
		if !isValid {
			return ctrl.Result{}, fmt.Errorf("the request does not match a cluster scoped access policy")
		}
	} else {
		var nsPolicies v1alpha1.JITAccessPolicyList
		listOpts := []client.ListOption{client.InNamespace(ns)}
		if err := r.List(ctx, &nsPolicies, listOpts...); err != nil {
			return ctrl.Result{}, err
		}

		isValid, matched = policy.IsRequestValid(obj, nsPolicies.Items)
		if !isValid {
			return ctrl.Result{}, fmt.Errorf("the request does not match a namespace scoped access policy")
		}
	}

	if matched == nil {
		return ctrl.Result{}, fmt.Errorf("the matched policy should not be nil")
	}

	if matched.RequiredApprovals != status.ApprovalsRequired {
		status.ApprovalsRequired = matched.RequiredApprovals
	}

	// State machine
	switch status.State {
	case v1alpha1.RequestStateApproved:
		result, err := r.handleApproved(ctx, obj, status)
		return result, err

	case v1alpha1.RequestStatePending:
		result, err := r.handlePending(ctx, obj, matched, status)
		return result, err
	}

	return ctrl.Result{}, nil
}

func (r *GenericJITAccessReconciler) handleApproved(
	ctx context.Context,
	obj common.JITAccessRequestObject,
	status *v1alpha1.JITAccessRequestStatus,
) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	spec := obj.GetSpec()

	// Already granted and expired
	if status.ExpiresAt != nil && time.Now().After(status.ExpiresAt.Time) {
		if err := r.cleanupResources(ctx, obj); err != nil {
			log.Error(err, "an error occurred running cleanup for the expired request", "name", obj.GetName())
			return ctrl.Result{}, err
		}
		log.Info("resources cleaned up for expired request, deleting the request", "name", obj.GetName())
		_ = r.Delete(ctx, obj)
		return ctrl.Result{}, nil
	}

	// Set expire time if not set
	if status.ExpiresAt == nil {
		expireTime := metav1.NewTime(time.Now().Add(time.Duration(spec.DurationSeconds) * time.Second))
		status.ExpiresAt = &expireTime
	}

	// Handle pre-defined role or adhoc permissions
	roleKind := obj.GetRoleKind()
	ns := obj.GetNamespace()
	reqName := fmt.Sprintf("jit-access-%s", status.RequestId)

	isClusterScoped := ns == ""

	// Pre-defined Role/ClusterRole
	if spec.Role != "" {
		isClusterRole := roleKind == v1alpha1.RoleKindClusterRole

		if err := r.createRoleBinding(ctx, obj, spec.Role, reqName, isClusterScoped, isClusterRole); err != nil && !errors.IsAlreadyExists(err) {
			log.Error(err, "an error occurred creating the role binding for the request", "name", obj.GetName(), "subject", spec.Subject, "roleKind", roleKind, "role", spec.Role)
			return ctrl.Result{}, err
		}
		status.RoleBindingCreated = true
		log.Info("Granted Role for request", "name", obj.GetName(), "subject", spec.Subject, "roleKind", roleKind, "role", spec.Role)
	}

	// Adhoc permissions
	if len(spec.Permissions) > 0 {
		adhocName := fmt.Sprintf("jit-access-adhoc-%s", status.RequestId)
		if err := r.createRole(ctx, obj, adhocName, spec.Permissions, isClusterScoped); err != nil && !errors.IsAlreadyExists(err) {
			log.Error(err, "an error occurred creating the adhoc role for the request", "name", obj.GetName(), "subject", spec.Subject, "roleKind", roleKind, "role", spec.Role)
			return ctrl.Result{}, err
		}
		status.AdhocRoleCreated = true

		if err := r.createRoleBinding(ctx, obj, adhocName, adhocName, isClusterScoped, isClusterRole); err != nil && !errors.IsAlreadyExists(err) {
			log.Error(err, "an error occurred creating the adhoc role binding for the request", "name", obj.GetName(), "subject", spec.Subject, "roleKind", roleKind, "role", spec.Role)
			return ctrl.Result{}, err
		}
		status.AdhocRoleBindingCreated = true
	}

	return ctrl.Result{RequeueAfter: time.Duration(spec.DurationSeconds) * time.Second}, nil
}

func (r *GenericJITAccessReconciler) handlePending(
	ctx context.Context,
	obj common.JITAccessRequestObject,
	matchedPolicy *v1alpha1.SubjectPolicy,
	status *v1alpha1.JITAccessRequestStatus,
) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	approved := set.New[string]()
	denied := set.New[string]()

	// Fetch responses
	if obj.GetNamespace() != "" {
		// Namespaced responses
		responses := &v1alpha1.JITAccessResponseList{}
		if err := r.List(ctx, responses, client.InNamespace(obj.GetNamespace()), client.MatchingFields{"spec.requestRef": obj.GetName()}); err != nil {
			log.Error(err, "an error occurred fetching responses for the request", "name", obj.GetName())
			return ctrl.Result{}, err
		}
		for _, resp := range responses.Items {
			switch resp.Spec.Response {
			case v1alpha1.ResponseStateApproved:
				approved.Insert(resp.Spec.Approver)
			case v1alpha1.ResponseStateDenied:
				denied.Insert(resp.Spec.Approver)
			}
		}
	} else {
		// Cluster-scoped responses
		responses := &v1alpha1.ClusterJITAccessResponseList{}
		if err := r.List(ctx, responses, client.MatchingFields{"spec.requestRef": obj.GetName()}); err != nil {
			log.Error(err, "an error occurred fetching responses for the request", "name", obj.GetName())
			return ctrl.Result{}, err
		}
		for _, resp := range responses.Items {
			switch resp.Spec.Response {
			case v1alpha1.ResponseStateApproved:
				approved.Insert(resp.Spec.Approver)
			case v1alpha1.ResponseStateDenied:
				denied.Insert(resp.Spec.Approver)
			}
		}
	}

	if approved.Len() != status.ApprovalsReceived {
		status.ApprovalsReceived = approved.Len()
	}

	if denied.Len() > 0 {
		status.State = v1alpha1.RequestStateDenied
	} else if approved.Len() >= matchedPolicy.RequiredApprovals {
		status.State = v1alpha1.RequestStateApproved
	}

	if status.State == v1alpha1.RequestStateApproved {
		return r.handleApproved(ctx, obj, status)
	}

	return ctrl.Result{}, nil
}

func (r *GenericJITAccessReconciler) cleanupResources(ctx context.Context, obj common.JITAccessRequestObject) error {
	log := logf.FromContext(ctx)
	status := obj.GetStatus()
	ns := obj.GetNamespace()
	requestId := status.RequestId

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
	if status.RoleBindingCreated {
		key := client.ObjectKey{Name: fmt.Sprintf("jit-access-%s", requestId)}
		var rb client.Object
		if ns == "" {
			rb = &rbacv1.ClusterRoleBinding{}
		} else {
			rb = &rbacv1.RoleBinding{}
			key.Namespace = ns
		}
		deleteResource(key, rb, "RoleBinding/ClusterRoleBinding")
	}

	// Adhoc RoleBinding / ClusterRoleBinding
	if status.AdhocRoleBindingCreated {
		key := client.ObjectKey{Name: fmt.Sprintf("jit-access-adhoc-%s", requestId)}
		var rb client.Object
		if ns == "" {
			rb = &rbacv1.ClusterRoleBinding{}
		} else {
			rb = &rbacv1.RoleBinding{}
			key.Namespace = ns
		}
		deleteResource(key, rb, "Adhoc RoleBinding/ClusterRoleBinding")
	}

	// Adhoc Role / ClusterRole
	if status.AdhocRoleCreated {
		key := client.ObjectKey{Name: fmt.Sprintf("jit-access-adhoc-%s", requestId)}
		var roleObj client.Object
		if ns == "" {
			roleObj = &rbacv1.ClusterRole{}
		} else {
			roleObj = &rbacv1.Role{}
			key.Namespace = ns
		}
		deleteResource(key, roleObj, "Adhoc Role/ClusterRole")
	}

	// Delete all responses
	var responses []client.Object
	if ns != "" {
		responseList := &v1alpha1.JITAccessResponseList{}
		if err := r.List(ctx, responseList, client.InNamespace(ns), client.MatchingFields{"spec.requestRef": obj.GetName()}); err != nil {
			errs = append(errs, fmt.Errorf("failed to list JITAccessResponses: %w", err))
		} else {
			for i := range responseList.Items {
				responses = append(responses, &responseList.Items[i])
			}
		}
	} else {
		responseList := &v1alpha1.ClusterJITAccessResponseList{}
		if err := r.List(ctx, responseList, client.MatchingFields{"spec.requestRef": obj.GetName()}); err != nil {
			errs = append(errs, fmt.Errorf("failed to list ClusterJITAccessResponses: %w", err))
		} else {
			for i := range responseList.Items {
				responses = append(responses, &responseList.Items[i])
			}
		}
	}

	for _, resp := range responses {
		if err := r.Delete(ctx, resp); err != nil {
			errs = append(errs, fmt.Errorf("failed to delete response %s: %w", resp.GetName(), err))
		} else {
			log.Info("Deleted response", "name", resp.GetName())
		}
	}

	if len(errs) > 0 {
		return goerr.Join(errs...)
	}

	return nil
}

func (r *GenericJITAccessReconciler) ensureFinalizer(ctx context.Context, obj client.Object, finalizer string) error {
	if obj.GetDeletionTimestamp().IsZero() && !controllerutil.ContainsFinalizer(obj, finalizer) {
		patch := client.MergeFrom(obj.DeepCopyObject().(client.Object))
		controllerutil.AddFinalizer(obj, finalizer)
		if err := r.Patch(ctx, obj, patch); err != nil {
			return fmt.Errorf("failed to add finalizer: %w", err)
		}
	}
	return nil
}

func (r *GenericJITAccessReconciler) removeFinalizer(ctx context.Context, obj client.Object, finalizer string) error {
	if controllerutil.ContainsFinalizer(obj, finalizer) {
		patch := client.MergeFrom(obj.DeepCopyObject().(client.Object))
		controllerutil.RemoveFinalizer(obj, finalizer)
		return r.Patch(ctx, obj, patch)
	}
	return nil
}

func (r *GenericJITAccessReconciler) createRole(
	ctx context.Context,
	obj common.JITAccessRequestObject,
	name string,
	rules []rbacv1.PolicyRule,
	clusterRole bool,
) error {
	var role client.Object
	if clusterRole {
		role = &rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Rules:      rules,
		}
	} else {
		role = &rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: obj.GetNamespace()},
			Rules:      rules,
		}
	}
	return r.Create(ctx, role)
}

func (r *GenericJITAccessReconciler) createRoleBinding(
	ctx context.Context,
	obj common.JITAccessRequestObject,
	roleName string,
	bindingName string,
	clusterScoped bool,
	clusterRole bool,
) error {
	subject := rbacv1.Subject{Kind: "User", Name: obj.GetSpec().Subject, APIGroup: "rbac.authorization.k8s.io"}
	if clusterScoped {
		rb := &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: bindingName},
			Subjects:   []rbacv1.Subject{subject},
			RoleRef:    rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: roleName},
		}
		return r.Create(ctx, rb)
	} else {
		kind := utils.Ternary(clusterRole, "ClusterRole", "Role")

		rb := &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: bindingName, Namespace: obj.GetNamespace()},
			Subjects:   []rbacv1.Subject{subject},
			RoleRef:    rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: kind, Name: roleName},
		}
		return r.Create(ctx, rb)
	}
}
