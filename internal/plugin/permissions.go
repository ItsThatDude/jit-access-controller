package plugin

import (
	"fmt"
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"
)

// parsePermissions parses "verbs:resources" strings into PolicyRule objects
func parsePermissions(perms []string) []rbacv1.PolicyRule {
	var rules []rbacv1.PolicyRule
	for _, p := range perms {
		parts := strings.Split(p, ":")
		if len(parts) != 2 {
			fmt.Printf("invalid permission format: %s (expected verbs:resources)\n", p)
			continue
		}
		verbs := strings.Split(parts[0], ",")
		resources := strings.Split(parts[1], ",")
		rules = append(rules, rbacv1.PolicyRule{
			Verbs:     verbs,
			Resources: resources,
		})
	}
	return rules
}
