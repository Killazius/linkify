package random

import (
	"math/rand"
	"strings"
	"time"
)

//nolint:gosec // Weak random number generator is acceptable for non-cryptographic use cases.
var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

const alphabet = "123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"

func NewRandomString(size int) string {
	var (
		builder strings.Builder
	)

	for i := 0; i < size; i++ {
		builder.WriteString(string(alphabet[rnd.Intn(len(alphabet))]))
	}
	return builder.String()
}
