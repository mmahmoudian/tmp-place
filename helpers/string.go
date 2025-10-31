package helpers

import (
	"math/rand"
	"time"
)

// GenerateRandomTag returns a random alphanumeric string of the given length.
// If length is zero, it defaults to 6.
// The generated string includes uppercase letters, lowercase letters, and digits.
func GenerateRandomTag(length ...int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	n := 6
	if len(length) > 0 {
		n = length[0]
	}
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
