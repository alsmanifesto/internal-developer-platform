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
	"time"
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

// ContainerStartedAt returns the time the named container was started.
// Returns a zero time.Time (not an error) if the container is not found or not running.
func ContainerStartedAt(containerName string) (time.Time, error) {
	cmd := exec.Command("docker", "inspect",
		"--format", "{{.State.StartedAt}}",
		containerName,
	)
	out, err := cmd.Output()
	if err != nil {
		// Container doesn't exist or isn't running — not a hard error
		return time.Time{}, nil
	}
	t, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(string(out)))
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing container start time: %w", err)
	}
	return t, nil
}

// ComposeDown stops and removes containers, networks, and images created by
// ComposeUp for the given compose directory.
// Equivalent to: docker compose down --rmi all --volumes --remove-orphans
func ComposeDown(composeDir string) error {
	cmd := exec.Command("docker", "compose", "down",
		"--rmi", "all",
		"--volumes",
		"--remove-orphans",
	)
	cmd.Dir = composeDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker compose down: %w", err)
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
