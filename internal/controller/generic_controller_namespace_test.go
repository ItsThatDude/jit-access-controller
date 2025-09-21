package controller

import (
	"context"
	"fmt"
	"time"

	"antware.xyz/jitaccess/api/v1alpha1"
	"antware.xyz/jitaccess/internal/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("GenericJITAccessReconciler with envtest", func() {
	var (
		ctx        context.Context
		reconciler *GenericJITAccessReconciler
		policyObj  *v1alpha1.JITAccessPolicy
	)

	BeforeEach(func() {
		ctx = context.Background()

		// Create policy object with unique name per run
		policyName := fmt.Sprintf("test-policy-%d", time.Now().UnixNano())
		policyObj = &v1alpha1.JITAccessPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      policyName,
				Namespace: "default",
			},
			Spec: v1alpha1.JITAccessPolicySpec{
				Policies: []v1alpha1.SubjectPolicy{
					{
						Subjects:           []string{"user1"},
						RequiredApprovals:  1,
						AllowedRoles:       []string{"edit"},
						Approvers:          []string{"admin"},
						MaxDurationSeconds: 3600,
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, policyObj)).To(Succeed())
		waitForCreated(ctx, k8sClient, client.ObjectKeyFromObject(policyObj), &v1alpha1.JITAccessPolicy{})

		reconciler = &GenericJITAccessReconciler{
			Client: mgr.GetClient(),
			Scheme: scheme.Scheme,
		}
	})

	AfterEach(func() {
		// Clean up policy
		Expect(k8sClient.Delete(ctx, policyObj)).To(Succeed())
		waitForDeleted(ctx, k8sClient, client.ObjectKeyFromObject(policyObj), &v1alpha1.JITAccessPolicy{})
	})

	It("should create a pending JITAccessRequest and update status", func() {
		requestName := fmt.Sprintf("test-pending-request-%d", time.Now().UnixNano())
		requestObj := &v1alpha1.JITAccessRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      requestName,
				Namespace: "default",
			},
			Spec: v1alpha1.JITAccessRequestSpec{
				JITAccessRequestBaseSpec: v1alpha1.JITAccessRequestBaseSpec{
					Subject:         "user1",
					Role:            "edit",
					DurationSeconds: 300,
					Justification:   "test",
				},
			},
		}
		Expect(k8sClient.Create(ctx, requestObj)).To(Succeed())
		waitForCreated(ctx, k8sClient, client.ObjectKeyFromObject(requestObj), &v1alpha1.JITAccessRequest{})

		// Reconcile once to process status/finalizer
		reconcileOnce(ctx, reconciler, client.ObjectKeyFromObject(requestObj))

		type requestStatus struct {
			ID         string
			State      v1alpha1.RequestState
			Finalizers []string
		}

		Eventually(func() (requestStatus, error) {
			updated := &v1alpha1.JITAccessRequest{}
			err := k8sClient.Get(ctx, client.ObjectKeyFromObject(requestObj), updated)
			if err != nil {
				return requestStatus{}, err
			}
			return requestStatus{
				ID:         updated.Status.RequestId,
				State:      updated.Status.State,
				Finalizers: updated.Finalizers,
			}, nil
		}, 5*time.Second, 100*time.Millisecond).Should(SatisfyAll(
			WithTransform(func(rs requestStatus) string { return rs.ID }, Not(BeEmpty())),
			WithTransform(func(rs requestStatus) v1alpha1.RequestState { return rs.State }, Equal(v1alpha1.RequestStatePending)),
			WithTransform(func(rs requestStatus) []string { return rs.Finalizers }, ContainElement(common.JITFinalizer)),
		))

		// Cleanup
		Expect(k8sClient.Delete(ctx, requestObj)).To(Succeed())
		waitForMarkedForDeletion(ctx, k8sClient, client.ObjectKeyFromObject(requestObj), &v1alpha1.JITAccessRequest{})

		reconcileOnce(ctx, reconciler, client.ObjectKeyFromObject(requestObj))
		waitForDeleted(ctx, k8sClient, client.ObjectKeyFromObject(requestObj), &v1alpha1.JITAccessRequest{})
	})

	It("should cleanup deleted JITAccessRequest and remove finalizer", func() {
		requestName := fmt.Sprintf("test-delete-request-%d", time.Now().UnixNano())
		requestObj := &v1alpha1.JITAccessRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      requestName,
				Namespace: "default",
				Finalizers: []string{
					common.JITFinalizer,
				},
			},
			Spec: v1alpha1.JITAccessRequestSpec{
				JITAccessRequestBaseSpec: v1alpha1.JITAccessRequestBaseSpec{
					Subject:         "user1",
					Role:            "edit",
					DurationSeconds: 300,
				},
			},
		}
		Expect(k8sClient.Create(ctx, requestObj)).To(Succeed())
		waitForCreated(ctx, k8sClient, client.ObjectKeyFromObject(requestObj), &v1alpha1.JITAccessRequest{})

		// Delete the object (simulate user deletion)
		Expect(k8sClient.Delete(ctx, requestObj)).To(Succeed())
		waitForMarkedForDeletion(ctx, k8sClient, client.ObjectKeyFromObject(requestObj), &v1alpha1.JITAccessRequest{})

		// Reconcile to handle finalizer cleanup
		reconcileOnce(ctx, reconciler, client.ObjectKeyFromObject(requestObj))

		// Wait until fully deleted
		waitForDeleted(ctx, k8sClient, client.ObjectKeyFromObject(requestObj), &v1alpha1.JITAccessRequest{})
	})
})
