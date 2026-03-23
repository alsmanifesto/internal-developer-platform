// Package detector identifies the technology stack from a project's Dockerfile.
package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Stack represents a detected technology stack.
type Stack string

const (
	StackGo      Stack = "go"
	StackPython  Stack = "python"
	StackSpark   Stack = "spark"
	StackKafka   Stack = "kafka"
	StackUnknown Stack = "unknown"
)

// DetectStack reads the Dockerfile in the given project path and returns the
// detected stack. Returns StackUnknown (not an error) when no rule matches.
func DetectStack(projectPath string) (Stack, error) {
	dockerfilePath := filepath.Join(projectPath, "Dockerfile")

	data, err := os.ReadFile(dockerfilePath)
	if err != nil {
		return StackUnknown, fmt.Errorf("reading Dockerfile: %w", err)
	}

	content := strings.ToLower(string(data))

	switch {
	// Spark must be checked before Python because spark images also contain "python"
	case strings.Contains(content, "spark"):
		return StackSpark, nil
	case strings.Contains(content, "kafka"):
		return StackKafka, nil
	case strings.Contains(content, "golang"):
		return StackGo, nil
	case strings.Contains(content, "python"):
		return StackPython, nil
	default:
		return StackUnknown, nil
	}
}
