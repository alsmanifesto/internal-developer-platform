// Package utils provides path and file validation helpers.
package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ValidatePath checks that the given path exists and contains a Dockerfile.
func ValidatePath(path string) error {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("path %q does not exist", path)
	}
	if err != nil {
		return fmt.Errorf("accessing path %q: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path %q is not a directory", path)
	}

	dockerfilePath := filepath.Join(path, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); errors.Is(err, os.ErrNotExist) {
		return errors.New("Dockerfile not found in provided path. A valid project must include a Dockerfile.")
	}

	return nil
}

// AbsPath returns the absolute version of path.
func AbsPath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolving absolute path for %q: %w", path, err)
	}
	return abs, nil
}
