package shared

import (
	"io"
	"os"
	"testing"
)

// captureStdout runs fn while redirecting stdout to a pipe and returns the captured output.
func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = old
		_ = r.Close()
		_ = w.Close()
	}()

	fn()
	_ = w.Close() // close writer so ReadAll completes
	b, _ := io.ReadAll(r)
	return string(b)
}

func TestMsg_KnownColor_PipedOutputsPlain(t *testing.T) {
	// When stdout is a pipe (as in this capture), Msg should not emit color codes
	out := captureStdout(func() {
		Msg("red", "hello")
	})
	// Note: current implementation prints two args to Println, yielding a trailing space
	expected := "hello \n"
	if out != expected {
		t.Fatalf("Msg red piped output = %q; want %q", out, expected)
	}
}

func TestMsg_CaseInsensitive_Piped(t *testing.T) {
	out := captureStdout(func() {
		Msg("ReD", "world")
	})
	expected := "world \n"
	if out != expected {
		t.Fatalf("Msg case-insensitive piped output = %q; want %q", out, expected)
	}
}

func TestMsg_UnknownColor_Piped(t *testing.T) {
	out := captureStdout(func() {
		Msg("unknown", "plain")
	})
	expected := "plain \n"
	if out != expected {
		t.Fatalf("Msg unknown color piped output = %q; want %q", out, expected)
	}
}

func TestMsg_WithArgs_Piped(t *testing.T) {
	out := captureStdout(func() {
		Msg("red", "hello ", "X", 1)
	})
	// Println separates args by a space, plus the message has a trailing space
	expected := "hello  X1\n"
	if out != expected {
		t.Fatalf("Msg with args piped = %q; want %q", out, expected)
	}
}

func TestMsgf_FormatsAndAddsNewline_DefaultColor(t *testing.T) {
	out := captureStdout(func() {
		Msgf("", "value=%d %s", 42, "ok")
	})
	expected := "value=42 ok\n"
	if out != expected {
		t.Fatalf("Msgf output = %q; want %q", out, expected)
	}
}
