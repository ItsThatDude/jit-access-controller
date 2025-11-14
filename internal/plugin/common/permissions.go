package common

import (
	"fmt"
	"log"
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"
)

// parsePermissions parses strings in the format:
// verbs:resource[.group][/subresource]
// Examples:
//
//	get:pods
//	get,list:deployments.apps/status
//	create:foos.example.com
func ParsePermissions(perms []string) []rbacv1.PolicyRule {
	rules := make([]rbacv1.PolicyRule, 0, len(perms))

	for _, p := range perms {
		parts := strings.Split(p, ":")
		if len(parts) != 2 {
			log.Printf("invalid permission format: %s (expected verbs:resources)\n", p)
			continue
		}

		verbs := strings.Split(parts[0], ",")
		rawResources := strings.Split(parts[1], ",")

		var resources []string
		var apiGroups []string

		for _, r := range rawResources {
			res := r
			group := ""

			// Handle subresource (pods/status)
			var sub string
			if strings.Contains(res, "/") {
				tokens := strings.SplitN(res, "/", 2)
				res = tokens[0]
				sub = tokens[1]
			}

			// Handle API group (deployments.apps)
			if strings.Contains(res, ".") {
				tokens := strings.SplitN(res, ".", 2)
				res = tokens[0]
				group = tokens[1]
			}

			// Combine resource + subresource
			if sub != "" {
				res = fmt.Sprintf("%s/%s", res, sub)
			}

			resources = append(resources, res)
			apiGroups = append(apiGroups, group)
		}

		rule := rbacv1.PolicyRule{
			Verbs:     verbs,
			Resources: resources,
		}

		// Only set API groups if at least one is non-empty
		hasGroups := false
		for _, g := range apiGroups {
			if g != "" {
				hasGroups = true
				break
			}
		}
		if hasGroups {
			rule.APIGroups = apiGroups
		}

		rules = append(rules, rule)
	}

	return rules
}
