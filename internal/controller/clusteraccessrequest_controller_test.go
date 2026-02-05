/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/itsthatdude/jit-access-controller/api/v1alpha1"
	common "github.com/itsthatdude/jit-access-controller/internal/common"
	"github.com/itsthatdude/jit-access-controller/internal/policy"
	"github.com/itsthatdude/jit-access-controller/internal/processors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

var _ = Describe("ClusterAccessRequest Controller", func() {
	var (
		ctx        context.Context
		reconciler *ClusterAccessRequestReconciler
		policyObj  *v1alpha1.ClusterAccessPolicy
	)

	BeforeEach(func() {
		ctx = context.Background()

		// Create policy object with unique name per run
		policyName := fmt.Sprintf("test-policy-%d", time.Now().UnixNano())
		policyObj = &v1alpha1.ClusterAccessPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: policyName,
			},
			Spec: v1alpha1.ClusterAccessPolicySpec{
				SubjectPolicy: v1alpha1.SubjectPolicy{
					Requesters:        []rbacv1.Subject{{Kind: rbacv1.UserKind, Name: "user1"}},
					RequiredApprovals: 1,
					AllowedRoles:      []rbacv1.RoleRef{{APIGroup: "rbac.authorization.k8s.io", Kind: common.RoleKindCluster, Name: "edit"}},
					Approvers:         []rbacv1.Subject{{Kind: rbacv1.UserKind, Name: "admin"}},
					MaxDuration:       "60m",
				},
			},
		}

		reconciler = &ClusterAccessRequestReconciler{
			Client:         mgr.GetClient(),
			Scheme:         scheme.Scheme,
			PolicyManager:  policy.NewPolicyManager(),
			PolicyResolver: &policy.PolicyResolver{},
		}

		reconciler.Processor = &processors.RequestProcessor{
			Client:         reconciler.Client,
			Scheme:         reconciler.Scheme,
			PolicyManager:  reconciler.PolicyManager,
			PolicyResolver: reconciler.PolicyResolver,
		}

		reconciler.PolicyManager.Update([]common.AccessPolicyObject{policyObj})
	})

	AfterEach(func() {
	})

	It("should fail to reconcile ClusterAccessRequest", func() {
		requestName := fmt.Sprintf("test-approve-request-%d", time.Now().UnixNano())

		requestObj := &v1alpha1.ClusterAccessRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name: requestName,
			},
			Spec: v1alpha1.ClusterAccessRequestSpec{
				AccessRequestBaseSpec: v1alpha1.AccessRequestBaseSpec{
					Subject: "user1",
					Role:    rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: common.RoleKindCluster, Name: "no-policy"},
					// nolint:goconst
					Duration:      "10m",
					Justification: "test",
				},
			},
		}

		Expect(k8sClient.Create(ctx, requestObj)).To(Succeed())
		waitForCreated(ctx, k8sClient, client.ObjectKeyFromObject(requestObj), requestObj)
		reconcileOnce(ctx, reconciler, client.ObjectKeyFromObject(requestObj)).ShouldNot(Succeed())
	})

	It("should create grant for approved ClusterAccessRequest", func() {
		requestName := fmt.Sprintf("test-approve-request-%d", time.Now().UnixNano())

		requestObj := &v1alpha1.ClusterAccessRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name: requestName,
			},
			Spec: v1alpha1.ClusterAccessRequestSpec{
				AccessRequestBaseSpec: v1alpha1.AccessRequestBaseSpec{
					Subject: "user1",
					Role:    rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: common.RoleKindCluster, Name: "edit"},
					// nolint:goconst
					Duration:      "10m",
					Justification: "test",
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
		responseObj := &v1alpha1.ClusterAccessResponse{
			ObjectMeta: metav1.ObjectMeta{
				Name: responseName,
			},
			Spec: v1alpha1.AccessResponseSpec{
				RequestRef: requestName,
				Approver:   "admin",
				Response:   v1alpha1.ResponseStateApproved,
			},
		}

		// Create the response and wait for it to be created
		Expect(k8sClient.Create(ctx, responseObj)).To(Succeed())
		waitForCreated(ctx, k8sClient, client.ObjectKeyFromObject(responseObj), &v1alpha1.ClusterAccessResponse{})

		// Reconcile the request again, to process the response
		reconcileOnce(ctx, reconciler, client.ObjectKeyFromObject(requestObj)).Should(Succeed())

		// Wait for the GrantCreated status to be set
		Eventually(func() bool {
			_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(requestObj), requestObj)
			return requestObj.Status.GrantCreated
		}, 5*time.Second, 100*time.Millisecond).Should(BeTrue())

		// Wait for the Grant to be created
		waitForCreated(ctx, k8sClient, client.ObjectKey{Name: requestName}, &v1alpha1.ClusterAccessGrant{})

		// Delete the object (simulate user deletion)
		Expect(k8sClient.Delete(ctx, requestObj)).To(Succeed())
		waitForDeletionTimestamp(ctx, k8sClient, client.ObjectKeyFromObject(requestObj), &v1alpha1.ClusterAccessRequest{})

		// Reconcile to handle finalizer cleanup
		reconcileOnce(ctx, reconciler, client.ObjectKeyFromObject(requestObj)).Should(Succeed())

		// Wait until fully deleted
		waitForDeleted(ctx, k8sClient, client.ObjectKeyFromObject(requestObj), &v1alpha1.ClusterAccessRequest{})
	})
})
