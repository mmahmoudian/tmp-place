package server

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"strings"
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

// SanitizeInput removes leading and trailing whitespace from the input string.
// Only allows safe characters for sql (alphanumeric and some special characters)
// and returns the sanitized string.
func SanitizeInput(input string) string {
	// Remove leading and trailing whitespace
	input = strings.TrimSpace(input)

	// Allow only safe characters (alphanumeric and some special characters)
	var safeChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_."
	for _, char := range input {
		if !strings.ContainsRune(safeChars, char) {
			input = strings.ReplaceAll(input, string(char), "")
		}
	}
	return input
}

// PrepareSecret sanitizes and truncates the secret to a maximum length of 42 characters.
// It uses SanitizeInput to clean the input string and then truncates it if necessary.
// Returns the prepared secret string.
func PrepareSecret(secret string) string {
	sanitized := SanitizeInput(secret)
	// calculate sha256 hash of the secret
	hash := sha256.Sum256([]byte(sanitized))
	return fmt.Sprintf("%x", hash)
}
