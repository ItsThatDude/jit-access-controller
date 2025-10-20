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

type RequestReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	SystemNamespace string
}

func (r *RequestReconciler) SetupWithManagerCluster(mgr ctrl.Manager) error {
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
		Named("request-reconciler-cluster").
		Complete(r)
}

func (r *RequestReconciler) SetupWithManagerNamespaced(mgr ctrl.Manager) error {
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
		Named("request-reconciler-namespaced").
		Complete(r)
}

func (r *RequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if req.Namespace == "" {
		var clusterObj v1alpha1.ClusterJITAccessRequest
		err := r.Get(ctx, types.NamespacedName{Name: req.Name}, &clusterObj)
		if err != nil {
			if errors.IsNotFound(err) {
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, err
		}
		return r.reconcileRequest(ctx, &clusterObj)
	}

	var nsObj v1alpha1.JITAccessRequest
	err := r.Get(ctx, req.NamespacedName, &nsObj)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	return r.reconcileRequest(ctx, &nsObj)
}

func (r *RequestReconciler) reconcileRequest(ctx context.Context, obj common.JITAccessRequestObject) (ctrl.Result, error) {
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

	// Set request expire time if not set
	if status.RequestExpiresAt == nil {
		expireTime := metav1.NewTime(time.Now().Add(time.Duration(60) * time.Minute))
		status.RequestExpiresAt = &expireTime
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

	if status.State != v1alpha1.RequestStateApproved &&
		status.RequestExpiresAt != nil && time.Now().After(status.RequestExpiresAt.Time) {
		status.State = v1alpha1.RequestStateExpired
	}

	if status.State == v1alpha1.RequestStateExpired {
		return r.handleExpired(ctx, obj)
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

	if status.State == v1alpha1.RequestStatePending {
		return r.handlePending(ctx, obj, matched, status)
	}

	return ctrl.Result{}, nil
}

func (r *RequestReconciler) handleApproved(
	ctx context.Context,
	obj common.JITAccessRequestObject,
	status *v1alpha1.JITAccessRequestStatus,
	approvers []string,
) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	spec := obj.GetSpec()

	if err := r.createGrant(ctx, obj, approvers); err != nil && !errors.IsAlreadyExists(err) {
		log.Error(err, "an error occurred creating the access grant for the request", "name", obj.GetName(), "subject", spec.Subject, common.RoleKindRole, spec.Role)
		return ctrl.Result{}, err
	}

	status.GrantCreated = true

	duration, err := time.ParseDuration(spec.Duration)
	if err != nil {
		log.Error(err, "failed to parse duration string", "duration", duration)
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: duration * time.Second}, nil
}

func (r *RequestReconciler) handlePending(
	ctx context.Context,
	obj common.JITAccessRequestObject,
	matchedPolicy *v1alpha1.SubjectPolicy,
	status *v1alpha1.JITAccessRequestStatus,
) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	spec := obj.GetSpec()

	approved := set.New[string]()
	denied := set.New[string]()

	// Fetch responses
	if obj.GetNamespace() == "" {
		// Cluster-scoped responses
		responses := &v1alpha1.ClusterJITAccessResponseList{}
		if err := r.List(ctx, responses, client.MatchingFields{"spec.requestRef": obj.GetName()}); err != nil {
			log.Error(err, "an error occurred fetching responses for the request", "name", obj.GetName())
			return ctrl.Result{}, err
		}
		for _, resp := range responses.Items {
			if resp.Spec.Approver != spec.Subject {
				switch resp.Spec.Response {
				case v1alpha1.ResponseStateApproved:
					approved.Insert(resp.Spec.Approver)
				case v1alpha1.ResponseStateDenied:
					denied.Insert(resp.Spec.Approver)
				}
			}
		}
	} else {
		// Namespaced responses
		responses := &v1alpha1.JITAccessResponseList{}
		if err := r.List(ctx, responses, client.InNamespace(obj.GetNamespace()), client.MatchingFields{"spec.requestRef": obj.GetName()}); err != nil {
			log.Error(err, "an error occurred fetching responses for the request", "name", obj.GetName())
			return ctrl.Result{}, err
		}
		for _, resp := range responses.Items {
			if resp.Spec.Approver != spec.Subject {
				switch resp.Spec.Response {
				case v1alpha1.ResponseStateApproved:
					approved.Insert(resp.Spec.Approver)
				case v1alpha1.ResponseStateDenied:
					denied.Insert(resp.Spec.Approver)
				}
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
		return r.handleApproved(ctx, obj, status, approved.UnsortedList())
	}

	if status.RequestExpiresAt != nil {
		return ctrl.Result{RequeueAfter: time.Until(status.RequestExpiresAt.Time)}, nil
	}

	return ctrl.Result{}, nil
}

func (r *RequestReconciler) handleExpired(
	ctx context.Context,
	obj common.JITAccessRequestObject,
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

func (r *RequestReconciler) cleanupResources(ctx context.Context, obj common.JITAccessRequestObject) error {
	log := logf.FromContext(ctx)
	ns := obj.GetNamespace()

	var errs []error

	// Delete all responses
	var responses []client.Object
	if ns == "" {
		responseList := &v1alpha1.ClusterJITAccessResponseList{}
		if err := r.List(ctx, responseList, client.MatchingFields{"spec.requestRef": obj.GetName()}); err != nil {
			errs = append(errs, fmt.Errorf("failed to list ClusterJITAccessResponses: %w", err))
		} else {
			for i := range responseList.Items {
				responses = append(responses, &responseList.Items[i])
			}
		}
	} else {
		responseList := &v1alpha1.JITAccessResponseList{}
		if err := r.List(ctx, responseList, client.InNamespace(ns), client.MatchingFields{"spec.requestRef": obj.GetName()}); err != nil {
			errs = append(errs, fmt.Errorf("failed to list JITAccessResponses: %w", err))
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

func (r *RequestReconciler) ensureFinalizer(ctx context.Context, obj client.Object, finalizer string) error {
	if obj.GetDeletionTimestamp().IsZero() && !controllerutil.ContainsFinalizer(obj, finalizer) {
		patch := client.MergeFrom(obj.DeepCopyObject().(client.Object))
		controllerutil.AddFinalizer(obj, finalizer)
		if err := r.Patch(ctx, obj, patch); err != nil {
			return fmt.Errorf("failed to add finalizer: %w", err)
		}
	}
	return nil
}

func (r *RequestReconciler) removeFinalizer(ctx context.Context, obj client.Object, finalizer string) error {
	if controllerutil.ContainsFinalizer(obj, finalizer) {
		patch := client.MergeFrom(obj.DeepCopyObject().(client.Object))
		controllerutil.RemoveFinalizer(obj, finalizer)
		return r.Patch(ctx, obj, patch)
	}
	return nil
}

func (r *RequestReconciler) createGrant(
	ctx context.Context,
	obj common.JITAccessRequestObject,
	approvers []string,
) error {
	reqName := obj.GetName()
	spec := obj.GetSpec()
	status := obj.GetStatus()
	ns := obj.GetNamespace()

	isClusterGrant := ns == ""

	labels := common.CommonLabels()

	grant := &v1alpha1.JITAccessGrant{
		ObjectMeta: metav1.ObjectMeta{Namespace: r.SystemNamespace, Name: reqName, Labels: labels},
		Spec:       v1alpha1.JITAccessGrantSpec{},
	}

	if err := r.Create(ctx, grant); err != nil {
		return err
	}

	original := grant.DeepCopy()

	grant.Status = v1alpha1.JITAccessGrantStatus{
		Request:   reqName,
		RequestId: status.RequestId,

		Subject:    spec.Subject,
		ApprovedBy: approvers,

		Role:        spec.Role,
		Permissions: spec.Permissions,
		Duration:    spec.Duration,
	}

	if isClusterGrant {
		grant.Status.Scope = v1alpha1.GrantScopeCluster
	} else {
		grant.Status.Scope = v1alpha1.GrantScopeNamespace
		grant.Status.Namespace = ns
	}

	patch := client.MergeFrom(original)

	if err := r.Status().Patch(ctx, grant, patch); err != nil {
		return err
	}

	return nil
}
