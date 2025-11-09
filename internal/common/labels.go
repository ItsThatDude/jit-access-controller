package common

func CommonLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/managed-by": "jit-access",
		"app.kubernetes.io/created-by": "controller-manager",
	}
}
