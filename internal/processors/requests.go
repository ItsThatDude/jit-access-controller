package processors

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/itsthatdude/jit-access-controller/api/v1alpha1"
	common "github.com/itsthatdude/jit-access-controller/internal/common"
	"github.com/itsthatdude/jit-access-controller/internal/metrics"
	"github.com/itsthatdude/jit-access-controller/internal/policy"
	"github.com/itsthatdude/jit-access-controller/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	set "k8s.io/apimachinery/pkg/util/sets"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type RequestProcessor struct {
	client.Client
	Scheme         *runtime.Scheme
	PolicyManager  *policy.PolicyManager
	PolicyResolver *policy.PolicyResolver
}

func (r *RequestProcessor) ReconcileRequest(ctx context.Context, obj common.AccessRequestObject) (ctrl.Result, error) {
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
	if status.RequestExpiresAt.IsZero() {
		status.RequestExpiresAt = metav1.NewTime(time.Now().Add(time.Duration(60) * time.Minute))
	}

	// Handle deletion
	if !obj.GetDeletionTimestamp().IsZero() {
		if controllerutil.ContainsFinalizer(obj, common.JITFinalizer) {
			log.Info("Cleaning up resources for request", "name", obj.GetName())
			if err := r.cleanupResponses(ctx, obj); err != nil {
				log.Error(err, "an error occurred running cleanup for the request", "name", obj.GetName())
				return ctrl.Result{}, err
			}
			log.Info("Successfully cleaned up resources for request", "name", obj.GetName())

			metrics.RequestStatus.Delete(
				prometheus.Labels{
					"scope":            string(obj.GetScope()),
					"target_namespace": obj.GetNamespace(),
					"request":          obj.GetName(),
					"subject":          obj.GetSubject(),
				},
			)

			log.Info("Removing finalizer for request", "name", obj.GetName())
			if err := RemoveFinalizer(r.Client, ctx, obj, common.JITFinalizer); err != nil {
				if !k8serrors.IsNotFound(err) {
					log.Error(err, "an error occurred removing the request finalizer", "name", obj.GetName())
				}
				return ctrl.Result{}, nil
			}
			log.Info("Removed finalizer for request", "name", obj.GetName())
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer
	if obj.GetDeletionTimestamp().IsZero() {
		err := EnsureFinalizerExists(r.Client, ctx, obj, common.JITFinalizer)
		if err != nil {
			log.Error(err, "an error occurred adding the finalizer to the request", "name", obj.GetName())
			return ctrl.Result{}, err
		}
	}

	if status.State != v1alpha1.RequestStateApproved &&
		!status.RequestExpiresAt.IsZero() && time.Now().After(status.RequestExpiresAt.Time) {
		status.State = v1alpha1.RequestStateExpired
	}

	if status.State == v1alpha1.RequestStateExpired {
		err := r.handleExpired(ctx, obj)
		return ctrl.Result{}, err
	}

	// Match against policies
	var policies = r.PolicyManager.GetSnapshot()

	matched_policy := r.PolicyResolver.Resolve(obj, policies)
	if matched_policy == nil {
		return ctrl.Result{}, fmt.Errorf("the request does not match an access policy")
	}

	policyName := matched_policy.GetName()
	policySpec := matched_policy.GetPolicy()

	if status.ResolvedPolicy == "" {
		status.ResolvedPolicy = policyName
	}

	if policySpec.RequiredApprovals != status.ApprovalsRequired {
		status.ApprovalsRequired = policySpec.RequiredApprovals
	}

	if status.State == v1alpha1.RequestStatePending {
		return r.handlePending(ctx, obj, &policySpec, status)
	}

	/*
		if status.State == v1alpha1.RequestStateDenied {

		}
	*/

	return ctrl.Result{}, nil
}

func (r *RequestProcessor) handleApproved(
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

	if err := r.createGrant(ctx, obj, status, approvers); err != nil && !k8serrors.IsAlreadyExists(err) {
		log.Error(err, "an error occurred creating the access grant for the request", "name", obj.GetName(), "subject", spec.Subject, "role", spec.Role)
		return ctrl.Result{}, err
	}

	status.GrantCreated = true

	metrics.RequestsApproved.WithLabelValues(string(obj.GetScope()), obj.GetNamespace(), obj.GetSubject()).Inc()

	if spec.Role.Name != "" {
		metrics.RolesGranted.WithLabelValues(string(obj.GetScope()), obj.GetNamespace(), obj.GetSubject(), spec.Role.Kind, spec.Role.Name).Inc()
	}

	r.updateRequestStatusMetric(obj, v1alpha1.RequestStateApproved)

	if len(spec.Permissions) > 0 {
		r.recordPermissionMetrics(obj, spec.Permissions)
	}

	// Requeue just after access expiry to handle cleanup
	return ctrl.Result{RequeueAfter: duration + time.Second}, nil
}

func (r *RequestProcessor) handlePending(
	ctx context.Context,
	obj common.AccessRequestObject,
	matchedPolicy *v1alpha1.SubjectPolicy,
	status *v1alpha1.AccessRequestStatus,
) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	spec := obj.GetSpec()

	approved := set.New[string]()
	denied := set.New[string]()

	approvals := set.New[v1alpha1.AccessRequestApproval]()

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
					approvals.Insert(v1alpha1.AccessRequestApproval{
						Approver:   resp.Spec.Approver,
						ApprovedAt: resp.CreationTimestamp,
					})
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
					approvals.Insert(v1alpha1.AccessRequestApproval{
						Approver:   resp.Spec.Approver,
						ApprovedAt: resp.CreationTimestamp,
					})
				case v1alpha1.ResponseStateDenied:
					denied.Insert(resp.Spec.Approver)
				}
			}
		}
	}

	if approved.Len() != status.ApprovalsReceived {
		status.ApprovalsReceived = approved.Len()
		status.Approvals = approvals.UnsortedList()
	}

	if denied.Len() > 0 {
		status.State = v1alpha1.RequestStateDenied
	} else if approved.Len() >= matchedPolicy.RequiredApprovals {
		status.State = v1alpha1.RequestStateApproved
	}

	if status.State == v1alpha1.RequestStateApproved {
		return r.handleApproved(ctx, obj, status, approved.UnsortedList())
	}

	r.updateRequestStatusMetric(obj, status.State)

	if !status.RequestExpiresAt.IsZero() {
		return ctrl.Result{RequeueAfter: time.Until(status.RequestExpiresAt.Time)}, nil
	}

	return ctrl.Result{}, nil
}

func (r *RequestProcessor) handleExpired(
	ctx context.Context,
	obj common.AccessRequestObject,
) error {
	log := logf.FromContext(ctx)

	if err := r.cleanupResponses(ctx, obj); err != nil {
		log.Error(err, "an error occurred running cleanup for the expired request", "name", obj.GetName())
		return err
	}

	metrics.RequestStatus.Delete(
		prometheus.Labels{
			"scope":            string(obj.GetScope()),
			"target_namespace": obj.GetNamespace(),
			"request":          obj.GetName(),
			"subject":          obj.GetSubject(),
		},
	)

	log.Info("resources cleaned up for expired request, deleting the request", "name", obj.GetName())
	_ = r.Delete(ctx, obj)

	return nil
}

func (r *RequestProcessor) cleanupResponses(ctx context.Context, obj common.AccessRequestObject) error {
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

func (r *RequestProcessor) createGrant(
	ctx context.Context,
	obj common.AccessRequestObject,
	status *v1alpha1.AccessRequestStatus,
	approvers []string,
) error {
	reqName := obj.GetName()
	spec := obj.GetSpec()
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

func (r *RequestProcessor) updateRequestStatusMetric(obj common.AccessRequestObject, state v1alpha1.RequestState) {
	var metricValue float64
	switch state {
	case v1alpha1.RequestStateApproved:
		metricValue = metrics.MetricStateApproved
	case v1alpha1.RequestStateDenied:
		metricValue = metrics.MetricStateDenied
	default:
		metricValue = metrics.MetricStatePending
	}

	metrics.RequestStatus.WithLabelValues(
		string(obj.GetScope()),
		obj.GetNamespace(),
		obj.GetName(),
		obj.GetSubject(),
	).Set(metricValue)
}

func (r *RequestProcessor) recordPermissionMetrics(obj common.AccessRequestObject, permissions []rbacv1.PolicyRule) {
	scope := string(obj.GetScope())
	namespace := obj.GetNamespace()
	subject := obj.GetSubject()

	for _, perm := range permissions {
		apiGroups := perm.APIGroups
		if len(apiGroups) == 0 {
			apiGroups = []string{""}
		}

		resourceNames := perm.ResourceNames
		if len(resourceNames) == 0 {
			resourceNames = []string{""}
		}

		for _, apiGroup := range apiGroups {
			for _, resource := range perm.Resources {
				for _, verb := range perm.Verbs {
					for _, resourceName := range resourceNames {
						metrics.PermissionsGranted.WithLabelValues(
							scope,
							namespace,
							subject,
							apiGroup,
							resource,
							verb,
							resourceName,
						).Inc()
					}
				}
			}
		}
	}
}
