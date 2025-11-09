package utils

import (
	"fmt"
	"math/rand/v2"
)

func GenerateRandomId() string {
	return fmt.Sprintf("%08x%08x", rand.Uint32(), rand.Uint32())
}

func FormatServiceAccountName(name, namespace string) string {
	return fmt.Sprintf("system:serviceaccount:%s:%s", namespace, name)
}
