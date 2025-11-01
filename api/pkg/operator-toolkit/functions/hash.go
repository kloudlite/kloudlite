package functions

import (
	"crypto/sha256"
	"fmt"
)

func SHA256Sum(b []byte) string {
	hasher := sha256.New()
	hasher.Write(b)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
