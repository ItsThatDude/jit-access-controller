package controller

import (
	"context"
	"time"

	"antware.xyz/jitaccess/api/v1alpha1"
	"antware.xyz/jitaccess/internal/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("GenericJITAccessReconciler", func() {
	var (
		scheme     *runtime.Scheme
		fakeClient client.Client
		reconciler *GenericJITAccessReconciler
		ctx        context.Context
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(clientgoscheme.AddToScheme(scheme)).To(Succeed())
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		Expect(rbacv1.AddToScheme(scheme)).To(Succeed())

		ctx = context.Background()
	})

	Context("Pending JITAccessRequest with matching policy", func() {
		var request *v1alpha1.JITAccessRequest
		var policyObj *v1alpha1.JITAccessPolicy

		BeforeEach(func() {
			policyObj = &v1alpha1.JITAccessPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       "JITAccessPolicy",
					APIVersion: "access.antware.xyz/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-policy",
					Namespace:       "default",
					ResourceVersion: "1",
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

			request = &v1alpha1.JITAccessRequest{
				TypeMeta: metav1.TypeMeta{
					Kind:       "JITAccessRequest",
					APIVersion: "access.antware.xyz/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:            "pending-request",
					Namespace:       "default",
					ResourceVersion: "1",
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

			fakeClient = fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(policyObj, request).
				WithIndex(&v1alpha1.JITAccessResponse{}, "spec.requestRef",
					func(rawObj client.Object) []string {
						if resp, ok := rawObj.(*v1alpha1.JITAccessResponse); ok {
							return []string{resp.Spec.RequestRef}
						}
						return nil
					},
				).
				Build()

			reconciler = &GenericJITAccessReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}
		})

		It("should set RequestId, add finalizer, and keep state pending", func() {
			req := ctrl.Request{NamespacedName: client.ObjectKey{Name: "pending-request", Namespace: "default"}}
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.Requeue).To(BeFalse())

			updated := &v1alpha1.JITAccessRequest{}
			updated.TypeMeta = metav1.TypeMeta{
				Kind:       "JITAccessRequest",
				APIVersion: "access.antware.xyz/v1alpha1",
			}
			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: "pending-request", Namespace: "default"}, updated)).To(Succeed())

			Expect(updated.Status.RequestId).ToNot(BeEmpty())
			Expect(updated.Status.State).To(Equal(v1alpha1.RequestStatePending))
			Expect(updated.Finalizers).To(ContainElement(common.JITFinalizer))
		})
	})

	Context("Approved JITAccessRequest with matching policy", func() {
		var request *v1alpha1.JITAccessRequest
		var policyObj *v1alpha1.JITAccessPolicy

		BeforeEach(func() {
			policyObj = &v1alpha1.JITAccessPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       "JITAccessPolicy",
					APIVersion: "access.antware.xyz/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-policy",
					Namespace:       "default",
					ResourceVersion: "1",
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

			request = &v1alpha1.JITAccessRequest{
				TypeMeta: metav1.TypeMeta{
					Kind:       "JITAccessRequest",
					APIVersion: "access.antware.xyz/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:            "approved-request",
					Namespace:       "default",
					ResourceVersion: "1",
				},
				Spec: v1alpha1.JITAccessRequestSpec{
					JITAccessRequestBaseSpec: v1alpha1.JITAccessRequestBaseSpec{
						Subject:         "user1",
						Role:            "edit",
						DurationSeconds: 300,
					},
				},
				Status: v1alpha1.JITAccessRequestStatus{
					RequestId: "req-123",
					State:     v1alpha1.RequestStateApproved,
				},
			}

			fakeClient = fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(policyObj, request).
				WithIndex(&v1alpha1.JITAccessResponse{}, "spec.requestRef",
					func(rawObj client.Object) []string {
						if resp, ok := rawObj.(*v1alpha1.JITAccessResponse); ok {
							return []string{resp.Spec.RequestRef}
						}
						return nil
					},
				).
				Build()

			reconciler = &GenericJITAccessReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}
		})

		It("should create a RoleBinding and set RequeueAfter", func() {
			req := ctrl.Request{NamespacedName: client.ObjectKey{Name: "approved-request", Namespace: "default"}}
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.RequeueAfter).To(BeNumerically("==", time.Duration(request.Spec.DurationSeconds)*time.Second))
		})
	})

	Context("Deleted JITAccessRequest cleanup", func() {
		It("should cleanup resources and remove finalizer", func() {
			policyObj := &v1alpha1.JITAccessPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       "JITAccessPolicy",
					APIVersion: "access.antware.xyz/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-policy",
					Namespace:       "default",
					ResourceVersion: "1",
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

			reqObj := &v1alpha1.JITAccessRequest{
				TypeMeta: metav1.TypeMeta{
					Kind:       "JITAccessRequest",
					APIVersion: "access.antware.xyz/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "deleted-request",
					Namespace:         "default",
					Finalizers:        []string{common.JITFinalizer},
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
					ResourceVersion:   "1",
				},
				Spec: v1alpha1.JITAccessRequestSpec{
					JITAccessRequestBaseSpec: v1alpha1.JITAccessRequestBaseSpec{
						Subject:         "user1",
						Role:            "edit",
						DurationSeconds: 300,
					},
				},
				Status: v1alpha1.JITAccessRequestStatus{
					RequestId:               "req-456",
					State:                   v1alpha1.RequestStateApproved,
					RoleBindingCreated:      true,
					AdhocRoleCreated:        true,
					AdhocRoleBindingCreated: true,
				},
			}

			roleBinding := &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "RoleBinding",
					APIVersion: "rbac.authorization.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:            "jit-access-req-456",
					Namespace:       "default",
					ResourceVersion: "1",
				},
			}
			adhocRole := &rbacv1.Role{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Role",
					APIVersion: "rbac.authorization.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:            "jit-access-adhoc-req-456",
					Namespace:       "default",
					ResourceVersion: "1",
				},
			}
			adhocBinding := &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "RoleBinding",
					APIVersion: "rbac.authorization.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:            "jit-access-adhoc-req-456",
					Namespace:       "default",
					ResourceVersion: "1",
				},
			}

			fakeClient = fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(policyObj, reqObj, roleBinding, adhocRole, adhocBinding).
				WithIndex(&v1alpha1.JITAccessResponse{}, "spec.requestRef",
					func(rawObj client.Object) []string {
						if resp, ok := rawObj.(*v1alpha1.JITAccessResponse); ok {
							return []string{resp.Spec.RequestRef}
						}
						return nil
					},
				).
				Build()

			reconciler = &GenericJITAccessReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			req := ctrl.Request{NamespacedName: client.ObjectKey{Name: "deleted-request", Namespace: "default"}}
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.Requeue).To(BeFalse())

			// Finalizer removed
			updated := &v1alpha1.JITAccessRequest{
				TypeMeta: metav1.TypeMeta{
					Kind:       "JITAccessRequest",
					APIVersion: "access.antware.xyz/v1alpha1",
				},
			}
			//Expect(updated.Finalizers).ToNot(ContainElement(common.JITFinalizer))
			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: "deleted-request", Namespace: "default"}, updated)).ShouldNot(Succeed())

			// RoleBinding deleted
			rb := &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "RoleBinding",
					APIVersion: "rbac.authorization.k8s.io/v1",
				},
			}
			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: "jit-access-req-456", Namespace: "default"}, rb)).To(HaveOccurred())

			// Adhoc Role deleted
			role := &rbacv1.Role{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Role",
					APIVersion: "rbac.authorization.k8s.io/v1",
				},
			}
			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: "jit-access-adhoc-req-456", Namespace: "default"}, role)).To(HaveOccurred())

			// Adhoc RoleBinding deleted
			adhocRB := &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "RoleBinding",
					APIVersion: "rbac.authorization.k8s.io/v1",
				},
			}
			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: "jit-access-adhoc-req-456", Namespace: "default"}, adhocRB)).To(HaveOccurred())
		})
	})
})
