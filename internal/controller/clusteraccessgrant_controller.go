package controller

import (
	"context"
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/events"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	accessv1alpha1 "github.com/itsthatdude/jit-access-controller/api/v1alpha1"
	"github.com/itsthatdude/jit-access-controller/internal/processors"
)

// ClusterAccessGrantReconciler reconciles a ClusterAccessGrant object
type ClusterAccessGrantReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Recorder  events.EventRecorder
	Processor *processors.GrantProcessor
}

// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusteraccessgrants,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusteraccessgrants/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=access.antware.xyz,resources=clusteraccessgrants/finalizers,verbs=update

// +kubebuilder:rbac:groups=events.k8s.io,resources=events,verbs=create;patch

func (r *ClusterAccessGrantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	var obj accessv1alpha1.ClusterAccessGrant
	err := r.Get(ctx, types.NamespacedName{Name: req.Name}, &obj)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	return r.Processor.ReconcileGrant(ctx, &obj)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterAccessGrantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Processor = &processors.GrantProcessor{
		Client:   r.Client,
		Scheme:   r.Scheme,
		Recorder: r.Recorder,
	}

	ctx := context.Background()
	indexer := mgr.GetFieldIndexer()

	if err := indexer.IndexField(ctx, &accessv1alpha1.ClusterAccessGrant{}, "status.requestId",
		func(obj client.Object) []string {
			if grant, ok := obj.(*accessv1alpha1.ClusterAccessGrant); ok {
				if grant.Status.RequestId == "" {
					return nil
				}
				return []string{grant.Status.RequestId}
			}
			return nil
		}); err != nil {
		return fmt.Errorf("failed to add index for requestId: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&accessv1alpha1.ClusterAccessGrant{}).
		Named("clustergrant-controller").
		Complete(r)
}
