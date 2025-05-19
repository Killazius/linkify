package random

import (
	"math/rand"
	"strings"
	"sync"
	"time"
)

const alphabet = "123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"

var (
	mu   sync.Mutex
	seed = time.Now().UnixNano()
)

func NewRandomString(size int) string {
	mu.Lock()
	//nolint:gosec // генератор случайных чисел приемлем для некриптографических случаев использования.
	rnd := rand.New(rand.NewSource(seed))
	seed++
	mu.Unlock()

	var builder strings.Builder
	builder.Grow(size)

	for i := 0; i < size; i++ {
		builder.WriteByte(alphabet[rnd.Intn(len(alphabet))])
	}
	return builder.String()
}
