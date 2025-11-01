package helpers

import (
	"strings"
	"testing"
)

func isAllowedCharset(s string) bool {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_."
	for _, r := range s {
		if !strings.ContainsRune(charset, r) {
			return false
		}
	}
	return true
}

func TestGenerateRandomTag_DefaultLength(t *testing.T) {
	tag := GenerateRandomTag()
	if len(tag) != 6 {
		t.Fatalf("expected default length 6, got %d (%q)", len(tag), tag)
	}
	if !isAllowedCharset(tag) {
		t.Fatalf("tag contains characters outside allowed charset: %q", tag)
	}
}

func TestGenerateRandomTag_CustomLength(t *testing.T) {
	for _, n := range []int{1, 8, 12, 32} {
		tag := GenerateRandomTag(n)
		if len(tag) != n {
			t.Errorf("expected length %d, got %d (%q)", n, len(tag), tag)
		}
		if !isAllowedCharset(tag) {
			t.Errorf("tag contains characters outside allowed charset: %q", tag)
		}
	}
}

func TestGenerateRandomTag_ZeroLength(t *testing.T) {
	tag := GenerateRandomTag(0)
	if tag != "" {
		t.Fatalf("expected empty string for zero length, got %q", tag)
	}
}

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		expected string
	}{
		{"trim and keep allowed", "  hello-world_123  ", "hello-world_123"},
		{"remove disallowed punctuation", "hi!@#$%^&*()=+[]{}|;:',<>?/~`", "hi"},
		{"remove spaces and punctuation", "bad chars: a b c", "badcharsabc"},
		{"non-ascii removed", "héllo世界", "hllo"},
		{"keep allowed set", ".-_Alpha123", ".-_Alpha123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeInput(tt.in)
			if got != tt.expected {
				t.Errorf("SanitizeInput(%q) = %q; want %q", tt.in, got, tt.expected)
			}
		})
	}
}

func TestPrepareSecret(t *testing.T) {
	// shorter than 42 and already clean
	if got := PrepareSecret("secret_value-._"); got != "secret_value-._" {
		t.Errorf("PrepareSecret clean short: got %q", got)
	}

	// sanitization removes disallowed
	if got := PrepareSecret("!!!secret$$$"); got != "secret" {
		t.Errorf("PrepareSecret sanitization: got %q; want %q", got, "secret")
	}

	// truncation to 42
	long := strings.Repeat("a", 50)
	got := PrepareSecret(long)
	if len(got) != 42 {
		t.Errorf("PrepareSecret length = %d; want 42", len(got))
	}
	expected := strings.Repeat("a", 42)
	if got != expected {
		t.Errorf("PrepareSecret value = %q; want %q", got, expected)
	}

	// sanitize then (no truncate since sanitized length < 42)
	messyLong := strings.Repeat("ab!@#", 20) // sanitizes to "ab" repeated 20 times = length 40
	got = PrepareSecret(messyLong)
	expectedSanitized := strings.Repeat("ab", 20)
	if got != expectedSanitized {
		t.Errorf("PrepareSecret sanitized = %q; want %q", got, expectedSanitized)
	}
	if !isAllowedCharset(got) {
		t.Errorf("PrepareSecret produced disallowed chars: %q", got)
	}
}
