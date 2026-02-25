package common

import (
	"reflect"
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"
)

func TestParsePermissions_SingleResourceNoGroup(t *testing.T) {
	input := []string{"get:pods"}
	got := ParsePermissions(input)

	if len(got) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(got))
	}

	want := rbacv1.PolicyRule{
		Verbs:     []string{"get"},
		Resources: []string{"pods"},
		APIGroups: []string{""},
	}

	if !reflect.DeepEqual(got[0].Verbs, want.Verbs) {
		t.Errorf("verbs mismatch: got %v want %v", got[0].Verbs, want.Verbs)
	}
	if !reflect.DeepEqual(got[0].Resources, want.Resources) {
		t.Errorf("resources mismatch: got %v want %v", got[0].Resources, want.Resources)
	}
	if !reflect.DeepEqual(got[0].APIGroups, want.APIGroups) {
		t.Errorf("apigroups mismatch: got %v want %v", got[0].APIGroups, want.APIGroups)
	}
}

func TestParsePermissions_GroupAndSubresourceAndMultipleResources(t *testing.T) {
	input := []string{"get,list:deployments.apps/status,services"}
	got := ParsePermissions(input)

	if len(got) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(got))
	}

	want := rbacv1.PolicyRule{
		Verbs:     []string{"get", "list"},
		Resources: []string{"deployments/status", "services"},
		// APIGroups should preserve per-resource groups: ["apps", ""]
		APIGroups: []string{"apps", ""},
	}

	if !reflect.DeepEqual(got[0].Verbs, want.Verbs) {
		t.Errorf("verbs mismatch: got %v want %v", got[0].Verbs, want.Verbs)
	}
	if !reflect.DeepEqual(got[0].Resources, want.Resources) {
		t.Errorf("resources mismatch: got %v want %v", got[0].Resources, want.Resources)
	}
	if !reflect.DeepEqual(got[0].APIGroups, want.APIGroups) {
		t.Errorf("apigroups mismatch: got %v want %v", got[0].APIGroups, want.APIGroups)
	}
}

func TestParsePermissions_CustomDomainGroup(t *testing.T) {
	input := []string{"create:foos.example.com"}
	got := ParsePermissions(input)

	if len(got) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(got))
	}

	want := rbacv1.PolicyRule{
		Verbs:     []string{"create"},
		Resources: []string{"foos"},
		APIGroups: []string{"example.com"},
	}

	if !reflect.DeepEqual(got[0].Verbs, want.Verbs) {
		t.Errorf("verbs mismatch: got %v want %v", got[0].Verbs, want.Verbs)
	}
	if !reflect.DeepEqual(got[0].Resources, want.Resources) {
		t.Errorf("resources mismatch: got %v want %v", got[0].Resources, want.Resources)
	}
	if !reflect.DeepEqual(got[0].APIGroups, want.APIGroups) {
		t.Errorf("apigroups mismatch: got %v want %v", got[0].APIGroups, want.APIGroups)
	}
}

func TestParsePermissions_InvalidEntrySkipped(t *testing.T) {
	input := []string{"badformat", "get:pods", "create:foos.example.com/status"}
	got := ParsePermissions(input)

	// badformat should be skipped; expect two rules parsed
	if len(got) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(got))
	}

	// First valid rule: get:pods
	want0 := rbacv1.PolicyRule{
		Verbs:     []string{"get"},
		Resources: []string{"pods"},
	}
	if !reflect.DeepEqual(got[0].Verbs, want0.Verbs) || !reflect.DeepEqual(got[0].Resources, want0.Resources) {
		t.Errorf("first rule mismatch: got %v/%v want %v/%v", got[0].Verbs, got[0].Resources, want0.Verbs, want0.Resources)
	}

	// Second valid rule: create:foos.example.com/status -> resource foos/status, group example.com
	want1 := rbacv1.PolicyRule{
		Verbs:     []string{"create"},
		Resources: []string{"foos/status"},
		APIGroups: []string{"example.com"},
	}
	if !reflect.DeepEqual(got[1].Verbs, want1.Verbs) {
		t.Errorf("second rule verbs mismatch: got %v want %v", got[1].Verbs, want1.Verbs)
	}
	if !reflect.DeepEqual(got[1].Resources, want1.Resources) {
		t.Errorf("second rule resources mismatch: got %v want %v", got[1].Resources, want1.Resources)
	}
	if !reflect.DeepEqual(got[1].APIGroups, want1.APIGroups) {
		t.Errorf("second rule apigroups mismatch: got %v want %v", got[1].APIGroups, want1.APIGroups)
	}
}
