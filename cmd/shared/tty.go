package shared

import "os"

// isStdinPiped reports whether stdin is coming from a pipe or redirection
// rather than an interactive terminal.
// It returns true when stdin is NOT a character device.
func isStdinPiped() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		// On error, assume interactive to avoid false positives
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) == 0
}

// isStdoutPiped reports whether stdout is being piped or redirected
// (i.e., not attached to an interactive terminal/TTY).
// It returns true when stdout is NOT a character device.
func isStdoutPiped() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		// On error, assume interactive to avoid false positives
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) == 0
}
