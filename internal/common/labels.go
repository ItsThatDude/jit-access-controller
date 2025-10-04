package common

func CommonLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/managed-by": "jitaccess",
		"app.kubernetes.io/created-by": "controller-manager",
	}
}
