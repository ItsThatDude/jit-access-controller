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
	"github.com/itsthatdude/jit-access-controller/internal/processors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

var _ = Describe("ClusterAccessGrant Controller", func() {
	var (
		reconciler *ClusterAccessGrantReconciler
	)
	Context("When reconciling a resource", func() {
		ctx := context.Background()

		BeforeEach(func() {
			reconciler = &ClusterAccessGrantReconciler{
				Client:   mgr.GetClient(),
				Scheme:   scheme.Scheme,
				Recorder: mgr.GetEventRecorderFor("accessgrant-controller"),
			}

			reconciler.Processor = &processors.GrantProcessor{
				Client:   reconciler.Client,
				Scheme:   reconciler.Scheme,
				Recorder: reconciler.Recorder,
			}
		})

		AfterEach(func() {

		})

		It("should create a Cluster Role Binding for approved ClusterAccessGrant", func() {
			grantName := fmt.Sprintf("test-grant-%d", time.Now().UnixNano())

			grantObj := &v1alpha1.ClusterAccessGrant{
				ObjectMeta: metav1.ObjectMeta{
					Name: grantName,
				},
			}

			Expect(k8sClient.Create(ctx, grantObj)).To(Succeed())
			waitForCreated(ctx, k8sClient, client.ObjectKeyFromObject(grantObj), grantObj)

			grantObj.Status.ApprovedBy = []string{"admin"}
			grantObj.Status.RequestId = grantName
			grantObj.Status.Role = rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: common.RoleKindCluster, Name: "edit"}
			grantObj.Status.Subject = "user1"
			// nolint:goconst
			grantObj.Status.Duration = "10m"

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
			))

			roleBindingName := fmt.Sprintf("jit-access-%s", grantObj.Status.RequestId)

			// Wait for the RoleBindingCreated status to be set
			Eventually(func() bool {
				_ = k8sClient.Get(ctx, client.ObjectKeyFromObject(grantObj), grantObj)
				return grantObj.Status.RoleBindingCreated
			}, 10*time.Second, 1*time.Second).Should(BeTrue())

			// See if the RoleBinding was actually created
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: roleBindingName}, &rbacv1.ClusterRoleBinding{})
				return err == nil
			}, 5*time.Second, 500*time.Millisecond).Should(BeTrue())

			// Delete the object (simulate user deletion)
			Expect(k8sClient.Delete(ctx, grantObj)).To(Succeed())

			// Reconcile to run the cleanup logic
			reconcileOnce(ctx, reconciler, client.ObjectKeyFromObject(grantObj)).Should(Succeed())

			Eventually(func() bool {
				rb := &rbacv1.ClusterRoleBinding{}
				err := k8sClient.Get(ctx, client.ObjectKey{Name: roleBindingName}, rb)
				fmt.Printf("CRB: %v\r\n", rb)
				return k8serrors.IsNotFound(err)
			}, 5*time.Second, 500*time.Millisecond).Should(BeTrue())

			Eventually(func() bool {
				grant := &v1alpha1.ClusterAccessGrant{}
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(grantObj), grant)
				return k8serrors.IsNotFound(err)
			}, 5*time.Second, 500*time.Millisecond).Should(BeTrue())
		})
	})
})
