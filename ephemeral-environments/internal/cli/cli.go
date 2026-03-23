// Package cli provides shared helpers for the ephemeral-env CLI.
package cli

import (
	"fmt"
	"os"
)

// Fatal prints a formatted error message to stderr and exits with code 1.
func Fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
