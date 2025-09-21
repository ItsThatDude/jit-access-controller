package controller

import (
	"context"
	"time"

	assert "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func waitForCreated(ctx context.Context, c client.Client, key client.ObjectKey, obj client.Object) {
	assert.Eventually(func() error { return c.Get(ctx, key, obj) }, 5*time.Second, 100*time.Millisecond).Should(assert.Succeed())
}

func waitForMarkedForDeletion(ctx context.Context, c client.Client, key client.ObjectKey, obj client.Object) {
	assert.Eventually(func() bool {
		err := c.Get(ctx, key, obj)
		return err == nil && obj.GetDeletionTimestamp() != nil
	}, 5*time.Second, 100*time.Millisecond).Should(assert.BeTrue())
}

func waitForDeleted(ctx context.Context, c client.Client, key client.ObjectKey, obj client.Object) {
	assert.Eventually(func() bool { return errors.IsNotFound(c.Get(ctx, key, obj)) }, 5*time.Second, 100*time.Millisecond).Should(assert.BeTrue())
}

func reconcileOnce(ctx context.Context, r *GenericJITAccessReconciler, key client.ObjectKey) {
	req := ctrl.Request{NamespacedName: key}
	assert.Eventually(func() error {
		_, err := r.Reconcile(ctx, req)
		return err
	}, 5*time.Second, 100*time.Millisecond).Should(assert.Succeed())
}
