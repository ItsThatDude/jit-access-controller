package policy

import (
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"
)

func TestRoleRefEquals(t *testing.T) {
	tests := []struct {
		name     string
		a, b     rbacv1.RoleRef
		expected bool
	}{
		{
			name:     "exact match",
			a:        rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: "admin"},
			b:        rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: "admin"},
			expected: true,
		},
		{
			name:     "empty APIGroup matches canonical",
			a:        rbacv1.RoleRef{APIGroup: "", Kind: "ClusterRole", Name: "admin"},
			b:        rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: "admin"},
			expected: true,
		},
		{
			name:     "different names",
			a:        rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: "admin"},
			b:        rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: "viewer"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := roleRefEquals(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFieldAllows(t *testing.T) {
	tests := []struct {
		name      string
		requested []string
		allowed   []string
		expected  bool
	}{
		{
			name:      "wildcard allows all",
			requested: []string{"pods"},
			allowed:   []string{"*"},
			expected:  true,
		},
		{
			name:      "exact match",
			requested: []string{"pods"},
			allowed:   []string{"pods", "services"},
			expected:  true,
		},
		{
			name:      "not allowed",
			requested: []string{"secrets"},
			allowed:   []string{"pods", "services"},
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fieldAllows(tt.requested, tt.allowed)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRoleRefSliceContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []rbacv1.RoleRef
		search   rbacv1.RoleRef
		expected bool
	}{
		{
			name: "found exact match",
			slice: []rbacv1.RoleRef{
				{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: "admin"},
			},
			search:   rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: "admin"},
			expected: true,
		},
		{
			name: "found with empty APIGroup",
			slice: []rbacv1.RoleRef{
				{APIGroup: "", Kind: "ClusterRole", Name: "admin"},
			},
			search:   rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: "admin"},
			expected: true,
		},
		{
			name:     "not found in empty slice",
			slice:    []rbacv1.RoleRef{},
			search:   rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: "admin"},
			expected: false,
		},
		{
			name: "not found in populated slice",
			slice: []rbacv1.RoleRef{
				{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: "viewer"},
			},
			search:   rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: "admin"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := roleRefSliceContains(tt.slice, tt.search)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRuleIsSubsetWithWildcard(t *testing.T) {
	tests := []struct {
		name      string
		requested rbacv1.PolicyRule
		allowed   rbacv1.PolicyRule
		expected  bool
	}{
		{
			name: "exact match",
			requested: rbacv1.PolicyRule{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get", "list"},
			},
			allowed: rbacv1.PolicyRule{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get", "list"},
			},
			expected: true,
		},
		{
			name: "requested subset of allowed",
			requested: rbacv1.PolicyRule{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get"},
			},
			allowed: rbacv1.PolicyRule{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get", "list", "watch"},
			},
			expected: true,
		},
		{
			name: "wildcard allows all",
			requested: rbacv1.PolicyRule{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments"},
				Verbs:     []string{"*"},
			},
			allowed: rbacv1.PolicyRule{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
			expected: true,
		},
		{
			name: "requested exceeds allowed",
			requested: rbacv1.PolicyRule{
				APIGroups: []string{""},
				Resources: []string{"pods", "secrets"},
				Verbs:     []string{"get"},
			},
			allowed: rbacv1.PolicyRule{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RuleIsSubsetWithWildcard(tt.requested, tt.allowed)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAllRequestedPolicyRulesAllowed(t *testing.T) {
	tests := []struct {
		name           string
		requestedRules []rbacv1.PolicyRule
		allowedRules   []rbacv1.PolicyRule
		expected       bool
	}{
		{
			name: "wildcard all requested rules allowed",
			requestedRules: []rbacv1.PolicyRule{
				{APIGroups: []string{""}, Resources: []string{"pods"}, Verbs: []string{"get"}},
				{APIGroups: []string{""}, Resources: []string{"services"}, Verbs: []string{"list"}},
			},
			allowedRules: []rbacv1.PolicyRule{
				{APIGroups: []string{"*"}, Resources: []string{"*"}, Verbs: []string{"*"}},
			},
			expected: true,
		},
		{
			name: "all requested rules allowed",
			requestedRules: []rbacv1.PolicyRule{
				{APIGroups: []string{""}, Resources: []string{"pods"}, Verbs: []string{"get"}},
				{APIGroups: []string{""}, Resources: []string{"services"}, Verbs: []string{"list"}},
			},
			allowedRules: []rbacv1.PolicyRule{
				{APIGroups: []string{""}, Resources: []string{"pods"}, Verbs: []string{"get"}},
				{APIGroups: []string{""}, Resources: []string{"services"}, Verbs: []string{"list"}},
			},
			expected: true,
		},
		{
			name: "one requested rule allowed",
			requestedRules: []rbacv1.PolicyRule{
				{APIGroups: []string{""}, Resources: []string{"pods"}, Verbs: []string{"get"}},
			},
			allowedRules: []rbacv1.PolicyRule{
				{APIGroups: []string{""}, Resources: []string{"pods"}, Verbs: []string{"get"}},
				{APIGroups: []string{""}, Resources: []string{"services"}, Verbs: []string{"list"}},
			},
			expected: true,
		},
		{
			name: "one rule not allowed",
			requestedRules: []rbacv1.PolicyRule{
				{APIGroups: []string{""}, Resources: []string{"pods"}, Verbs: []string{"get"}},
				{APIGroups: []string{""}, Resources: []string{"secrets"}, Verbs: []string{"get"}},
			},
			allowedRules: []rbacv1.PolicyRule{
				{APIGroups: []string{""}, Resources: []string{"pods"}, Verbs: []string{"get"}},
			},
			expected: false,
		},
		{
			name:           "empty requested rules always allowed",
			requestedRules: []rbacv1.PolicyRule{},
			allowedRules: []rbacv1.PolicyRule{
				{APIGroups: []string{""}, Resources: []string{"pods"}, Verbs: []string{"get"}},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AllRequestedPolicyRulesAllowed(tt.requestedRules, tt.allowedRules)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}
