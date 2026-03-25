package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/alsmanifesto/internal-developer-platform/ephemeral-env/internal/compose"
	"github.com/alsmanifesto/internal-developer-platform/ephemeral-env/internal/detector"
	"github.com/alsmanifesto/internal-developer-platform/ephemeral-env/internal/docker"
	"github.com/alsmanifesto/internal-developer-platform/ephemeral-env/internal/utils"
	"github.com/spf13/cobra"
)

var (
	flagPath    string
	flagEnvID   string
	flagDryRun  bool
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new ephemeral environment",
	Long: `Validates the project path, detects the stack, generates a docker-compose
file with Traefik labels, and brings the environment up.`,
	RunE: runCreate,
}

func init() {
	createCmd.Flags().StringVar(&flagPath, "path", "", "Path to the scaffold project folder (must contain a Dockerfile)")
	createCmd.Flags().StringVar(&flagEnvID, "env-id", "", "Unique environment identifier (e.g. payments-pr-123)")
	createCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Generate docker-compose.yml without running docker compose up")
	_ = createCmd.MarkFlagRequired("path")
	_ = createCmd.MarkFlagRequired("env-id")
}

func runCreate(cmd *cobra.Command, args []string) error {
	// 1. Validate path and Dockerfile
	if err := utils.ValidatePath(flagPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// 2. Detect stack
	stack, err := detector.DetectStack(flagPath)
	if err != nil {
		return fmt.Errorf("stack detection failed: %w", err)
	}

	fmt.Printf("⚡ Creating ephemeral environment %s (stack: %s)...\n", flagEnvID, stack)

	// 3. Warn if the project has no EXPOSE — URL won't serve traffic
	hasPort, err := detector.HasExposedPort(flagPath)
	if err != nil {
		return fmt.Errorf("checking exposed port: %w", err)
	}
	if !hasPort {
		fmt.Println()
		fmt.Println("⚠️  This project does not expose a port (no EXPOSE in Dockerfile).")
		fmt.Println("   It looks like a job or script, not a web service.")
		fmt.Println("   The environment will be created, but the preview URL will not return any response.")
		fmt.Println()
	}

	// 5. Resolve absolute path for compose context
	absPath, err := utils.AbsPath(flagPath)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	// 6. Generate docker-compose.yml
	composeDir, err := compose.Generate(flagEnvID, absPath)
	if err != nil {
		return fmt.Errorf("generating docker-compose: %w", err)
	}

	fmt.Printf("   docker-compose.yml written to %s/\n", composeDir)

	// 7. Skip execution if --dry-run
	if flagDryRun {
		fmt.Println()
		fmt.Println("ℹ️  Dry-run mode: docker compose up skipped.")
		fmt.Printf("   To start manually: docker compose -f %s/docker-compose.yml up -d --build\n", composeDir)
		return nil
	}

	// 8. Run docker compose up
	if err := docker.ComposeUp(composeDir); err != nil {
		switch {
		case errors.Is(err, docker.ErrBuildFailed):
			fmt.Fprintln(os.Stderr, "\n❌ Project has some bugs to fix in order to have a running environment.")
			fmt.Fprintln(os.Stderr, "   Review the build output above, fix the errors, and run again.")
		case errors.Is(err, docker.ErrNotRunning):
			fmt.Fprintln(os.Stderr, "\n⚠️  Project is not exposing a running service.")
			fmt.Fprintln(os.Stderr, "   The container started but exited immediately.")
			fmt.Fprintln(os.Stderr, "   This usually means the project is a function or job, not a long-running service.")
			fmt.Fprintf(os.Stderr, "   Inspect logs with: docker compose -f %s/docker-compose.yml logs\n", composeDir)
		default:
			fmt.Fprintf(os.Stderr, "\n❌ Failed to start environment: %v\n", err)
		}
		os.Exit(1)
	}

	// 9. Print result
	fmt.Println()
	fmt.Println("🌐 Preview environment ready:")
	fmt.Printf("   http://%s.local.scaffold.dev\n", flagEnvID)
	fmt.Println()
	fmt.Println("💰 Estimated cost rate: $0.20/hour")

	return nil
}
