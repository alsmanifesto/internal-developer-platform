// Package cli provides shared CLI utilities for the scaffold tool.
package cli

import (
	"fmt"
	"os"
)

// PrintError prints an error message to stderr and exits with code 1.
func PrintError(msg string, err error) {
	fmt.Fprintf(os.Stderr, "Error: %s: %v\n", msg, err)
	os.Exit(1)
}

// PrintSuccess prints a success message to stdout.
func PrintSuccess(msg string) {
	fmt.Println(msg)
}
