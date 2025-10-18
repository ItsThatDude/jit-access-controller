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

var _ = Describe("JITAccessReconciler with envtest", func() {
	var (
		ctx        context.Context
		reconciler *RequestReconciler
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
						AllowedRoles:       []rbacv1.RoleRef{{APIGroup: "rbac.authorization.k8s.io", Kind: common.RoleKindRole, Name: "edit"}},
						Approvers:          []string{"admin"},
						MaxDurationSeconds: 3600,
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, policyObj)).To(Succeed())
		waitForCreated(ctx, k8sClient, client.ObjectKeyFromObject(policyObj), &v1alpha1.JITAccessPolicy{})

		reconciler = &RequestReconciler{
			Client:          mgr.GetClient(),
			Scheme:          scheme.Scheme,
			SystemNamespace: "default",
		}
	})

	AfterEach(func() {
		// Clean up policy
		Expect(k8sClient.Delete(ctx, policyObj)).To(Succeed())
		waitForDeleted(ctx, k8sClient, client.ObjectKeyFromObject(policyObj), &v1alpha1.JITAccessPolicy{})
	})

	It("should create grant for approved JITAccessRequest", func() {
		requestName := fmt.Sprintf("test-approve-request-%d", time.Now().UnixNano())

		requestObj := &v1alpha1.JITAccessRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      requestName,
				Namespace: "default",
			},
			Spec: v1alpha1.JITAccessRequestSpec{
				JITAccessRequestBaseSpec: v1alpha1.JITAccessRequestBaseSpec{
					Subject:         "user1",
					Role:            rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: common.RoleKindRole, Name: "edit"},
					DurationSeconds: 300,
					Justification:   "test",
				},
			},
		}

		Expect(k8sClient.Create(ctx, requestObj)).To(Succeed())
		waitForCreated(ctx, k8sClient, client.ObjectKeyFromObject(requestObj), requestObj)
		reconcileOnce(ctx, reconciler, client.ObjectKeyFromObject(requestObj)).Should(Succeed())

		type requestStatus struct {
			ID         string
			State      v1alpha1.RequestState
			Finalizers []string
		}

		Eventually(func() (requestStatus, error) {
			err := k8sClient.Get(ctx, client.ObjectKeyFromObject(requestObj), requestObj)
			if err != nil {
				return requestStatus{}, err
			}
			return requestStatus{
				ID:         requestObj.Status.RequestId,
				State:      requestObj.Status.State,
				Finalizers: requestObj.Finalizers,
			}, nil
		}, 5*time.Second, 100*time.Millisecond).Should(SatisfyAll(
			WithTransform(func(rs requestStatus) string { return rs.ID }, Not(BeEmpty())),
			WithTransform(func(rs requestStatus) v1alpha1.RequestState { return rs.State }, Equal(v1alpha1.RequestStatePending)),
			WithTransform(func(rs requestStatus) []string { return rs.Finalizers }, ContainElement(common.JITFinalizer)),
		))

		responseName := fmt.Sprintf("test-approve-response-%d", time.Now().UnixNano())
		responseObj := &v1alpha1.JITAccessResponse{
			ObjectMeta: metav1.ObjectMeta{
				Name:      responseName,
				Namespace: requestObj.Namespace,
			},
			Spec: v1alpha1.JITAccessResponseSpec{
				RequestRef: requestName,
				Approver:   "admin",
				Response:   v1alpha1.ResponseStateApproved,
			},
		}

		// Create the response and wait for it to be created
		Expect(k8sClient.Create(ctx, responseObj)).To(Succeed())
		waitForCreated(ctx, k8sClient, client.ObjectKeyFromObject(responseObj), &v1alpha1.JITAccessResponse{})

		// Reconcile the request again, to process the response
		reconcileOnce(ctx, reconciler, client.ObjectKeyFromObject(requestObj)).Should(Succeed())

		// Wait for the GrantCreated status to be set
		Eventually(func() bool {
			_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(requestObj), requestObj)
			return requestObj.Status.GrantCreated
		}, 5*time.Second, 100*time.Millisecond).Should(BeTrue())

		// Wait for the Grant to be created
		waitForCreated(ctx, k8sClient, client.ObjectKey{Namespace: reconciler.SystemNamespace, Name: requestObj.Name}, &v1alpha1.JITAccessGrant{})

		// Delete the object (simulate user deletion)
		Expect(k8sClient.Delete(ctx, requestObj)).To(Succeed())
		waitForDeletionTimestamp(ctx, k8sClient, client.ObjectKeyFromObject(requestObj), &v1alpha1.JITAccessRequest{})

		// Reconcile to handle finalizer cleanup
		reconcileOnce(ctx, reconciler, client.ObjectKeyFromObject(requestObj)).Should(Succeed())

		// Wait until fully deleted
		waitForDeleted(ctx, k8sClient, client.ObjectKeyFromObject(requestObj), &v1alpha1.JITAccessRequest{})
	})
})
