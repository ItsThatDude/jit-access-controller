package processors

import (
	"context"

	common "github.com/itsthatdude/jit-access-controller/internal/common"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func EnsureFinalizerExists(c client.Client, ctx context.Context, obj client.Object, finalizer string) error {
	if !controllerutil.ContainsFinalizer(obj, common.JITFinalizer) {
		patch := client.MergeFrom(obj.DeepCopyObject().(client.Object))
		controllerutil.AddFinalizer(obj, common.JITFinalizer)
		if err := c.Patch(ctx, obj, patch); err != nil {
			return err
		}
	}
	return nil
}

func RemoveFinalizer(c client.Client, ctx context.Context, obj client.Object, finalizer string) error {
	if controllerutil.ContainsFinalizer(obj, finalizer) {
		patch := client.MergeFrom(obj.DeepCopyObject().(client.Object))
		controllerutil.RemoveFinalizer(obj, finalizer)
		if err := c.Patch(ctx, obj, patch); err != nil {
			return err
		}
	}
	return nil
}
