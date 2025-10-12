package controller

import (
	"context"
	"fmt"
	"time"

	"antware.xyz/jitaccess/api/v1alpha1"
	common "antware.xyz/jitaccess/internal/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("GenericJITGrantReconciler with envtest", func() {
	var (
		ctx        context.Context
		reconciler *GrantReconciler
	)

	BeforeEach(func() {
		ctx = context.Background()

		reconciler = &GrantReconciler{
			Client:          mgr.GetClient(),
			Scheme:          scheme.Scheme,
			SystemNamespace: "default",
		}
	})

	AfterEach(func() {

	})

	It("should create ClusterRoleBinding for approved cluster scoped JITAccessGrant", func() {
		grantName := fmt.Sprintf("test-grant-%d", time.Now().UnixNano())

		grantObj := &v1alpha1.JITAccessGrant{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: reconciler.SystemNamespace,
				Name:      grantName,
			},
			Spec: v1alpha1.JITAccessGrantSpec{},
		}

		Expect(k8sClient.Create(ctx, grantObj)).To(Succeed())
		waitForCreated(ctx, k8sClient, client.ObjectKeyFromObject(grantObj), grantObj)

		grantObj.Status.Scope = v1alpha1.GrantScopeCluster
		grantObj.Status.ApprovedBy = []string{"admin"}
		grantObj.Status.RequestId = "test-request"
		grantObj.Status.Role = rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: common.RoleKindCluster, Name: "edit"}
		grantObj.Status.Subject = "user1"
		grantObj.Status.DurationSeconds = 300

		Expect(k8sClient.Status().Update(ctx, grantObj)).To(Succeed())

		// Wait for the RequestId status to be set
		Eventually(func() bool {
			_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(grantObj), grantObj)
			return grantObj.Status.RequestId != ""
		}, 10*time.Second, 1*time.Second).Should(BeTrue())

		reconcileOnce(ctx, reconciler, client.ObjectKeyFromObject(grantObj)).Should(Succeed())

		type grantStatus struct {
			ID         string
			Finalizers []string
		}

		Eventually(func() (grantStatus, error) {
			err := k8sClient.Get(ctx, client.ObjectKeyFromObject(grantObj), grantObj)
			if err != nil {
				return grantStatus{}, err
			}
			return grantStatus{
				ID:         grantObj.Status.RequestId,
				Finalizers: grantObj.Finalizers,
			}, nil
		}, 5*time.Second, 500*time.Millisecond).Should(SatisfyAll(
			WithTransform(func(rs grantStatus) string { return rs.ID }, Not(BeEmpty())),
			WithTransform(func(rs grantStatus) []string { return rs.Finalizers }, ContainElement(common.JITFinalizer)),
		))

		// Wait for the RoleBindingCreated status to be set
		Eventually(func() bool {
			_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(grantObj), grantObj)
			return grantObj.Status.RoleBindingCreated
		}, 10*time.Second, 1*time.Second).Should(BeTrue())

		// See if the ClusterRoleBinding was actually created
		Eventually(func() bool {
			err := k8sClient.Get(ctx, client.ObjectKey{Name: grantObj.Name}, &rbacv1.ClusterRoleBinding{})
			return err != nil
		}, 5*time.Second, 500*time.Millisecond).Should(BeTrue())

		// Delete the object (simulate user deletion)
		Expect(k8sClient.Delete(ctx, grantObj)).To(Succeed())
		waitForDeletionTimestamp(ctx, k8sClient, client.ObjectKeyFromObject(grantObj), &v1alpha1.JITAccessGrant{})

		// Reconcile to handle finalizer cleanup
		reconcileOnce(ctx, reconciler, client.ObjectKeyFromObject(grantObj)).Should(Succeed())

		// Wait until fully deleted
		waitForDeleted(ctx, k8sClient, client.ObjectKeyFromObject(grantObj), &v1alpha1.JITAccessGrant{})
	})
})
