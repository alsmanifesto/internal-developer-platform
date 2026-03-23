// Package docker wraps docker CLI commands used by ephemeral-env.
package docker

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Sentinel errors returned by ComposeUp so callers can show targeted messages.
var (
	ErrBuildFailed  = errors.New("build failed")
	ErrNotRunning   = errors.New("container not running after start")
)

// ComposeUp runs `docker compose up -d --build` in the given directory.
// It streams output to the terminal while also capturing it to detect
// whether a failure is a build error or a runtime issue.
//
// Returns ErrBuildFailed if the image could not be built.
// Returns ErrNotRunning if the build succeeded but the container exited immediately.
func ComposeUp(composeDir string) error {
	var captured bytes.Buffer
	multi := io.MultiWriter(os.Stdout, &captured)

	cmd := exec.Command("docker", "compose", "up", "-d", "--build")
	cmd.Dir = composeDir
	cmd.Stdout = multi
	cmd.Stderr = multi

	if err := cmd.Run(); err != nil {
		if isBuildError(captured.String()) {
			return ErrBuildFailed
		}
		return fmt.Errorf("docker compose up: %w", err)
	}

	// Build succeeded — verify the app container is actually running.
	if !isContainerRunning(composeDir) {
		return ErrNotRunning
	}

	return nil
}

// isBuildError returns true when the captured docker output contains
// patterns emitted by BuildKit on a failed image build.
func isBuildError(output string) bool {
	markers := []string{
		"=> ERROR",
		"error building image",
		"failed to build",
		"failed to solve",
		"executor failed running",
	}
	lower := strings.ToLower(output)
	for _, m := range markers {
		if strings.Contains(lower, strings.ToLower(m)) {
			return true
		}
	}
	return false
}

// isContainerRunning checks whether at least one container managed by the
// compose project in composeDir has status "running".
func isContainerRunning(composeDir string) bool {
	cmd := exec.Command("docker", "compose", "ps", "--format", "{{.State}}")
	cmd.Dir = composeDir
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(out)), "running")
}
