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

package v1alpha1

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	accessv1alpha1 "antware.xyz/jitaccess/api/v1alpha1"
	"antware.xyz/jitaccess/internal/common"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("AccessRequest Webhook", func() {
	var (
		obj       *accessv1alpha1.AccessRequest
		oldObj    *accessv1alpha1.AccessRequest
		validator AccessRequestCustomValidator
	)

	BeforeEach(func() {
		obj = &accessv1alpha1.AccessRequest{}
		oldObj = &accessv1alpha1.AccessRequest{}
		validator = AccessRequestCustomValidator{
			client: k8sClient,
		}
		Expect(validator).NotTo(BeNil(), "Expected validator to be initialized")
		Expect(oldObj).NotTo(BeNil(), "Expected oldObj to be initialized")
		Expect(obj).NotTo(BeNil(), "Expected obj to be initialized")
	})

	AfterEach(func() {

	})

	Context("When creating or updating AccessRequest under Validating Webhook", func() {
		It("Should deny creation if a required field is missing", func() {
			By("simulating an invalid creation scenario")
			obj.Spec.Role = rbacv1.RoleRef{}
			Expect(validator.ValidateCreate(ctx, obj)).Error().To(HaveOccurred())
		})

		It("Should reject creation if request doesn't match policy", func() {
			By("simulating an invalid creation scenario")

			obj.Spec.Subject = "unknown_user"
			obj.Spec.Role = rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: common.RoleKindRole, Name: "edit"}
			obj.Spec.Duration = "5m"

			_, err := validator.ValidateCreate(ctx, obj)
			Expect(err).To(MatchError("access request did not match a policy"))
		})

		It("Should admit creation if all required fields are present", func() {
			By("simulating a valid creation scenario")
			// Create policy object with unique name per run
			policyName := fmt.Sprintf("test-policy-%d", time.Now().UnixNano())
			policyObj := &accessv1alpha1.AccessPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      policyName,
					Namespace: "default",
				},
				Spec: accessv1alpha1.AccessPolicySpec{
					SubjectPolicy: accessv1alpha1.SubjectPolicy{
						Subjects:          []string{"user1"},
						RequiredApprovals: 1,
						AllowedRoles:      []rbacv1.RoleRef{{APIGroup: "rbac.authorization.k8s.io", Kind: common.RoleKindRole, Name: "edit"}},
						Approvers:         []string{"admin"},
						MaxDuration:       "60m",
					},
				},
			}
			Expect(k8sClient.Create(ctx, policyObj)).To(Succeed())
			waitForCreated(ctx, k8sClient, client.ObjectKeyFromObject(policyObj), &accessv1alpha1.AccessPolicy{}).Should(Succeed())

			obj.Spec.Subject = "user1"
			obj.Spec.Role = rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: common.RoleKindRole, Name: "edit"}
			obj.Spec.Duration = "5m"

			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).ToNot(HaveOccurred())

			// Clean up the created policy
			Expect(k8sClient.Delete(ctx, policyObj)).To(Succeed())
			waitForDeleted(ctx, k8sClient, client.ObjectKeyFromObject(policyObj), &accessv1alpha1.AccessPolicy{}).Should(BeTrue())
		})
	})
})
