package shared

import (
	"os"
	"testing"
)

func TestIsStdinPiped_WithPipe(t *testing.T) {
	// Save and restore original Stdin
	old := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe error: %v", err)
	}
	defer func() {
		_ = r.Close()
		_ = w.Close()
		os.Stdin = old
	}()

	// Replace Stdin with the read end of a pipe
	os.Stdin = r
	// Write something so the pipe isn't at EOF
	_, _ = w.Write([]byte("data"))

	if got := isStdinPiped(); !got {
		t.Fatalf("isPiped() = false; want true when stdin is a pipe")
	}
}

func TestIsStdoutPiped_WithPipe(t *testing.T) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe error: %v", err)
	}
	defer func() {
		_ = r.Close()
		_ = w.Close()
		os.Stdout = old
	}()

	// Redirect stdout to the pipe
	os.Stdout = w

	if got := isStdoutPiped(); !got {
		t.Fatalf("isStdoutPiped() = false; want true when stdout is a pipe")
	}
}
