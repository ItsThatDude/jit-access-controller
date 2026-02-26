package policy

import (
	"context"
	"testing"

	accessv1alpha1 "github.com/itsthatdude/jit-access-controller/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestLoadPolicies(t *testing.T) {
	ctx := context.Background()

	// build a scheme containing both our APIs and the core k8s types (the
	// latter are needed by the fake client builder)
	sch := runtime.NewScheme()
	if err := scheme.AddToScheme(sch); err != nil {
		t.Fatalf("unable to add core scheme: %v", err)
	}
	if err := accessv1alpha1.AddToScheme(sch); err != nil {
		t.Fatalf("unable to add access scheme: %v", err)
	}

	// create one cluster-scoped and one namespaced policy object
	clPol := &accessv1alpha1.ClusterAccessPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "cluster-foo"},
		Spec: accessv1alpha1.ClusterAccessPolicySpec{
			SubjectPolicy: accessv1alpha1.SubjectPolicy{Priority: 5},
		},
	}

	nsPol := &accessv1alpha1.AccessPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "ns-bar", Namespace: "default"},
		Spec: accessv1alpha1.AccessPolicySpec{
			SubjectPolicy: accessv1alpha1.SubjectPolicy{Priority: 7},
		},
	}

	fakeClient := ctrlclient.NewClientBuilder().WithScheme(sch).
		WithObjects(clPol, nsPol).
		Build()

	// load into managers
	cm := NewPolicyManager()
	nm := NewPolicyManager()

	if err := LoadClusterPolicies(ctx, fakeClient, cm); err != nil {
		t.Fatalf("unexpected error loading cluster policies: %v", err)
	}
	if err := LoadNamespacedPolicies(ctx, fakeClient, nm); err != nil {
		t.Fatalf("unexpected error loading namespaced policies: %v", err)
	}

	snap := cm.GetSnapshot()
	if len(snap) != 1 {
		t.Fatalf("expected 1 cluster policy in snapshot, got %d", len(snap))
	}
	if snap[0].GetName() != clPol.Name {
		t.Errorf("unexpected cluster policy name: %s", snap[0].GetName())
	}

	snap = nm.GetSnapshot()
	if len(snap) != 1 {
		t.Fatalf("expected 1 namespaced policy in snapshot, got %d", len(snap))
	}
	if snap[0].GetName() != nsPol.Name {
		t.Errorf("unexpected namespaced policy name: %s", snap[0].GetName())
	}
}
