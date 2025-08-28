package utils

import (
	"fmt"
	"math/rand/v2"
)

func GenerateRandomId() string {
	return fmt.Sprintf("%08x", rand.Uint32())
}
