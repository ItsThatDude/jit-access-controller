package common

func CommonLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/managed-by": "kairos",
		"app.kubernetes.io/created-by": "controller-manager",
	}
}
