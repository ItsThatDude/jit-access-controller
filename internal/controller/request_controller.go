package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/itsthatdude/jitaccess-controller/api/v1alpha1"
	common "github.com/itsthatdude/jitaccess-controller/internal/common"
	"github.com/itsthatdude/jitaccess-controller/internal/metrics"
	"github.com/itsthatdude/jitaccess-controller/internal/policy"
	"github.com/itsthatdude/jitaccess-controller/internal/utils"
	"k8s.io/apimachinery/pkg/api/equality"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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

// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusteraccesspolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusteraccesspolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusteraccesspolicies/finalizers,verbs=update

// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusteraccessrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusteraccessrequests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusteraccessrequests/finalizers,verbs=update

// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusteraccessresponses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusteraccessresponses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusteraccessresponses/finalizers,verbs=update

// +kubebuilder:rbac:groups=access.antware.xyz,resources=accesspolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=access.antware.xyz,resources=accesspolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=access.antware.xyz,resources=accesspolicies/finalizers,verbs=update

// +kubebuilder:rbac:groups=access.antware.xyz,resources=accessrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=access.antware.xyz,resources=accessrequests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=access.antware.xyz,resources=accessrequests/finalizers,verbs=update

// +kubebuilder:rbac:groups=access.antware.xyz,resources=accessresponses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=access.antware.xyz,resources=accessresponses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=access.antware.xyz,resources=accessresponses/finalizers,verbs=update

// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;delete;bind;escalate

// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create;delete;bind;escalate

type RequestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *RequestReconciler) SetupWithManagerCluster(mgr ctrl.Manager) error {
	ctx := context.Background()
	indexer := mgr.GetFieldIndexer()

	if err := indexer.IndexField(ctx, &v1alpha1.ClusterAccessRequest{}, "status.requestId",
		func(obj client.Object) []string {
			if myObj, ok := obj.(*v1alpha1.ClusterAccessRequest); ok {
				if myObj.Status.RequestId == "" {
					return nil
				}
				return []string{myObj.Status.RequestId}
			}
			return nil
		}); err != nil {
		return fmt.Errorf("failed to add index for requestId: %w", err)
	}

	if err := indexer.IndexField(ctx, &v1alpha1.ClusterAccessResponse{}, "spec.requestRef",
		func(obj client.Object) []string {
			if myObj, ok := obj.(*v1alpha1.ClusterAccessResponse); ok {
				if myObj.Spec.RequestRef == "" {
					return nil
				}
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
		For(&v1alpha1.ClusterAccessRequest{}).
		Watches(
			&v1alpha1.ClusterAccessResponse{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				resp := obj.(*v1alpha1.ClusterAccessResponse)
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

	if err := indexer.IndexField(ctx, &v1alpha1.AccessRequest{}, "status.requestId",
		func(obj client.Object) []string {
			if myObj, ok := obj.(*v1alpha1.AccessRequest); ok {
				if myObj.Status.RequestId == "" {
					return nil
				}
				return []string{myObj.Status.RequestId}
			}
			return nil
		}); err != nil {
		return fmt.Errorf("failed to add index for requestId: %w", err)
	}

	if err := indexer.IndexField(ctx, &v1alpha1.AccessResponse{}, "spec.requestRef",
		func(obj client.Object) []string {
			if myObj, ok := obj.(*v1alpha1.AccessResponse); ok {
				if myObj.Spec.RequestRef == "" {
					return nil
				}
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
		For(&v1alpha1.AccessRequest{}).
		Watches(
			&v1alpha1.AccessResponse{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				resp := obj.(*v1alpha1.AccessResponse)
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
		var clusterObj v1alpha1.ClusterAccessRequest
		err := r.Get(ctx, types.NamespacedName{Name: req.Name}, &clusterObj)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, err
		}
		return r.reconcileRequest(ctx, &clusterObj)
	}

	var nsObj v1alpha1.AccessRequest
	err := r.Get(ctx, req.NamespacedName, &nsObj)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	return r.reconcileRequest(ctx, &nsObj)
}

func (r *RequestReconciler) reconcileRequest(ctx context.Context, obj common.AccessRequestObject) (ctrl.Result, error) {
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

		metrics.RequestsCreated.WithLabelValues(string(obj.GetScope()), obj.GetNamespace(), obj.GetSubject()).Inc()
	}

	// Set request expire time if not set
	if status.RequestExpiresAt == nil {
		expireTime := metav1.NewTime(time.Now().Add(time.Duration(60) * time.Minute))
		status.RequestExpiresAt = &expireTime
	}

	// Add finalizer
	if obj.GetDeletionTimestamp().IsZero() {
		if !controllerutil.ContainsFinalizer(obj, common.JITFinalizer) {
			patch := client.MergeFrom(obj.DeepCopyObject().(client.Object))
			controllerutil.AddFinalizer(obj, common.JITFinalizer)
			log.Info("Adding finalizer to request", "name", obj.GetName())
			if err := r.Patch(ctx, obj, patch); err != nil {
				return ctrl.Result{}, err
			}
			log.Info("Added finalizer to request", "name", obj.GetName())
		}
	}

	// Handle deletion
	if !obj.GetDeletionTimestamp().IsZero() {
		if controllerutil.ContainsFinalizer(obj, common.JITFinalizer) {
			log.Info("Cleaning up resources for request", "name", obj.GetName())
			if err := r.cleanupResources(ctx, obj); err != nil {
				log.Error(err, "an error occurred running cleanup for the request", "name", obj.GetName())
				return ctrl.Result{}, err
			}
			log.Info("Successfully cleaned up resources for request", "name", obj.GetName())

			log.Info("Removing finalizer for request", "name", obj.GetName())
			if err := r.removeFinalizer(ctx, obj, common.JITFinalizer); err != nil {
				log.Error(err, "an error occurred removing the request finalizer", "name", obj.GetName())
				return ctrl.Result{}, err
			}
			log.Info("Removed finalizer for request", "name", obj.GetName())
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
		var clusterPolicies v1alpha1.ClusterAccessPolicyList
		if err := r.List(ctx, &clusterPolicies); err != nil {
			return ctrl.Result{}, err
		}

		isValid, matched = policy.IsRequestValid(obj, clusterPolicies.Items)
		if !isValid {
			return ctrl.Result{}, fmt.Errorf("the request does not match a cluster scoped access policy")
		}
	} else {
		var nsPolicies v1alpha1.AccessPolicyList
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
	obj common.AccessRequestObject,
	status *v1alpha1.AccessRequestStatus,
	approvers []string,
) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	spec := obj.GetSpec()

	durationStr := spec.Duration
	if durationStr == "" {
		// nolint:goconst
		durationStr = "10m"
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		log.Error(err, "failed to parse duration string", "duration", durationStr)
		return ctrl.Result{}, err
	}

	if err := r.createGrant(ctx, obj, approvers); err != nil && !k8serrors.IsAlreadyExists(err) {
		log.Error(err, "an error occurred creating the access grant for the request", "name", obj.GetName(), "subject", spec.Subject, common.RoleKindRole, spec.Role)
		return ctrl.Result{}, err
	}

	status.GrantCreated = true

	metrics.RequestsApproved.WithLabelValues(string(obj.GetScope()), obj.GetNamespace(), obj.GetSubject()).Inc()

	if spec.Role.Name != "" {
		metrics.RolesGranted.WithLabelValues(string(obj.GetScope()), obj.GetNamespace(), obj.GetSubject(), spec.Role.Kind, spec.Role.Name).Inc()
	}

	// this may need to be revisited
	if len(spec.Permissions) > 0 {
		for _, perm := range spec.Permissions {
			for _, apiGroup := range perm.APIGroups {
				for _, resource := range perm.Resources {
					for _, verb := range perm.Verbs {
						metrics.PermissionsGranted.WithLabelValues(
							string(obj.GetScope()),
							obj.GetNamespace(),
							obj.GetSubject(),
							apiGroup,
							resource,
							verb,
						).Inc()
					}
				}
			}
		}
	}

	// Requeue just after access expiry to handle cleanup
	return ctrl.Result{RequeueAfter: duration + time.Second}, nil
}

func (r *RequestReconciler) handlePending(
	ctx context.Context,
	obj common.AccessRequestObject,
	matchedPolicy *v1alpha1.SubjectPolicy,
	status *v1alpha1.AccessRequestStatus,
) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	spec := obj.GetSpec()

	approved := set.New[string]()
	denied := set.New[string]()

	// Fetch responses
	if obj.GetScope() == v1alpha1.RequestScopeCluster {
		// Cluster-scoped responses
		responses := &v1alpha1.ClusterAccessResponseList{}
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
		responses := &v1alpha1.AccessResponseList{}
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
	obj common.AccessRequestObject,
) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	if err := r.cleanupResources(ctx, obj); err != nil {
		log.Error(err, "an error occurred running cleanup for the expired request", "name", obj.GetName())
		return ctrl.Result{RequeueAfter: 10 * time.Second}, err
	}

	log.Info("resources cleaned up for expired request, deleting the request", "name", obj.GetName())
	_ = r.Delete(ctx, obj)

	return ctrl.Result{}, nil
}

func (r *RequestReconciler) cleanupResources(ctx context.Context, obj common.AccessRequestObject) error {
	log := logf.FromContext(ctx)
	scope := obj.GetScope()
	ns := obj.GetNamespace()

	var errs []error

	// Delete all responses
	var responses []client.Object
	if scope == v1alpha1.RequestScopeCluster {
		responseList := &v1alpha1.ClusterAccessResponseList{}
		if err := r.List(ctx, responseList, client.MatchingFields{"spec.requestRef": obj.GetName()}); err != nil {
			errs = append(errs, fmt.Errorf("failed to list ClusterAccessResponses: %w", err))
		} else {
			for i := range responseList.Items {
				responses = append(responses, &responseList.Items[i])
			}
		}
	} else {
		responseList := &v1alpha1.AccessResponseList{}
		if err := r.List(ctx, responseList, client.InNamespace(ns), client.MatchingFields{"spec.requestRef": obj.GetName()}); err != nil {
			errs = append(errs, fmt.Errorf("failed to list AccessResponses: %w", err))
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
		return errors.Join(errs...)
	}

	return nil
}

func (r *RequestReconciler) removeFinalizer(ctx context.Context, obj client.Object, finalizer string) error {
	if controllerutil.ContainsFinalizer(obj, finalizer) {
		patch := client.MergeFrom(obj.DeepCopyObject().(client.Object))
		controllerutil.RemoveFinalizer(obj, finalizer)
		if err := r.Patch(ctx, obj, patch); err != nil {
			return err
		}
	}
	return nil
}

func (r *RequestReconciler) createGrant(
	ctx context.Context,
	obj common.AccessRequestObject,
	approvers []string,
) error {
	reqName := obj.GetName()
	spec := obj.GetSpec()
	status := obj.GetStatus()
	ns := obj.GetNamespace()

	isClusterGrant := obj.GetScope() == v1alpha1.RequestScopeCluster
	labels := common.CommonLabels()

	grantBaseStatus := v1alpha1.AccessGrantStatus{
		Request:   reqName,
		RequestId: status.RequestId,

		Subject:    spec.Subject,
		ApprovedBy: approvers,

		Role:        spec.Role,
		Permissions: spec.Permissions,
		Duration:    spec.Duration,
	}

	var grant common.AccessGrantObject

	if isClusterGrant {
		grant = &v1alpha1.ClusterAccessGrant{
			ObjectMeta: metav1.ObjectMeta{Name: reqName, Labels: labels},
		}
	} else {
		if spec.Role.Kind != common.RoleKindRole {
			return fmt.Errorf("invalid role kind for namespace scoped grant: %s", spec.Role.Kind)
		}

		grant = &v1alpha1.AccessGrant{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: reqName, Labels: labels},
		}
	}

	if err := controllerutil.SetControllerReference(obj, grant.(metav1.Object), r.Scheme); err != nil {
		return fmt.Errorf("failed to set owner reference on Grant %s: %w", grant.GetName(), err)
	}

	if err := r.Create(ctx, grant); err != nil {
		return err
	}

	original := grant.DeepCopyObject().(client.Object)
	grant.SetStatus(&grantBaseStatus)

	patch := client.MergeFrom(original)

	if err := r.Status().Patch(ctx, grant, patch); err != nil {
		return err
	}

	return nil
}
