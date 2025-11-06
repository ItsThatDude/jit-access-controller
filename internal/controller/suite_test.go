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
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	accessv1alpha1 "github.com/itsthatdude/jitaccess-controller/api/v1alpha1"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	ctx       context.Context
	cancel    context.CancelFunc
	testEnv   *envtest.Environment
	cfg       *rest.Config
	k8sClient client.Client
	mgr       ctrl.Manager
)

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	var err error
	err = accessv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = accessv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	// Retrieve the first found binary directory to allow running tests from IDEs
	if getFirstFoundEnvTestBinaryDir() != "" {
		testEnv.BinaryAssetsDirectory = getFirstFoundEnvTestBinaryDir()
	}

	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	mgr, err = ctrl.NewManager(cfg, ctrl.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())

	Expect(mgr.GetFieldIndexer().IndexField(ctx, &accessv1alpha1.AccessResponse{}, "spec.requestRef",
		func(obj client.Object) []string {
			if r, ok := obj.(*accessv1alpha1.AccessResponse); ok {
				return []string{r.Spec.RequestRef}
			}
			return nil
		})).To(Succeed())

	Expect(mgr.GetFieldIndexer().IndexField(ctx, &accessv1alpha1.ClusterAccessResponse{}, "spec.requestRef",
		func(obj client.Object) []string {
			if r, ok := obj.(*accessv1alpha1.ClusterAccessResponse); ok {
				return []string{r.Spec.RequestRef}
			}
			return nil
		})).To(Succeed())

	go func() {
		Expect(mgr.Start(ctx)).To(Succeed())
	}()

	cacheReady := mgr.GetCache().WaitForCacheSync(ctx)
	Expect(cacheReady).To(BeTrue())

	k8sClient = mgr.GetClient()
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancel()
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

// getFirstFoundEnvTestBinaryDir locates the first binary in the specified path.
// ENVTEST-based tests depend on specific binaries, usually located in paths set by
// controller-runtime. When running tests directly (e.g., via an IDE) without using
// Makefile targets, the 'BinaryAssetsDirectory' must be explicitly configured.
//
// This function streamlines the process by finding the required binaries, similar to
// setting the 'KUBEBUILDER_ASSETS' environment variable. To ensure the binaries are
// properly set up, run 'make setup-envtest' beforehand.
func getFirstFoundEnvTestBinaryDir() string {
	basePath := filepath.Join("..", "..", "bin", "k8s")
	entries, err := os.ReadDir(basePath)
	if err != nil {
		logf.Log.Error(err, "Failed to read directory", "path", basePath)
		return ""
	}
	for _, entry := range entries {
		if entry.IsDir() {
			return filepath.Join(basePath, entry.Name())
		}
	}
	return ""
}

func waitForCreated(ctx context.Context, c client.Client, key client.ObjectKey, obj client.Object) {
	Eventually(func() error { return c.Get(ctx, key, obj) }, 5*time.Second, 100*time.Millisecond).Should(Succeed())
}

func waitForDeletionTimestamp(ctx context.Context, c client.Client, key client.ObjectKey, obj client.Object) {
	Eventually(func() bool {
		err := c.Get(ctx, key, obj)
		return err == nil && obj.GetDeletionTimestamp() != nil
	}, 5*time.Second, 100*time.Millisecond).Should(BeTrue())
}

func waitForDeleted(ctx context.Context, c client.Client, key client.ObjectKey, obj client.Object) {
	Eventually(func() bool { return errors.IsNotFound(c.Get(ctx, key, obj)) }, 5*time.Second, 100*time.Millisecond).Should(BeTrue())
}

func reconcileOnce(ctx context.Context, r reconcile.TypedReconciler[reconcile.Request], key client.ObjectKey) AsyncAssertion {
	req := ctrl.Request{NamespacedName: key}
	return Eventually(func() error {
		_, err := r.Reconcile(ctx, req)
		return err
	}, 5*time.Second, 100*time.Millisecond)
}
