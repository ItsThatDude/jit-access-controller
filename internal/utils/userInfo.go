package utils

import (
	"fmt"

	authenticationv1 "k8s.io/api/authentication/v1"
)

func IsController(systemNamespace string, serviceAccount string, user authenticationv1.UserInfo) bool {
	if serviceAccount == "" {
		return false
	}

	serviceAccountName := fmt.Sprintf("system:serviceaccount:%s:%s", systemNamespace, serviceAccount)
	groupName := fmt.Sprintf("system:serviceaccounts:%s", systemNamespace)

	isControllerUser := user.Username == serviceAccountName

	if !isControllerUser {
		return false
	}

	for _, g := range user.Groups {
		if g == groupName {
			return true
		}
	}

	return false
}
