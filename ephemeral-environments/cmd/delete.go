package cmd

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/alsmanifesto/internal-developer-platform/ephemeral-env/internal/docker"
	"github.com/spf13/cobra"
)

var flagDeleteEnvID string

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an ephemeral environment",
	Long: `Stops and removes all Docker resources for the given environment:
containers, images, volumes, and networks created by ephemeral-env create.
Also removes the local envs/<env-id>/ directory.`,
	RunE: runDelete,
}

func init() {
	deleteCmd.Flags().StringVar(&flagDeleteEnvID, "env-id", "", "Environment identifier to delete (e.g. payments-pr-123)")
	_ = deleteCmd.MarkFlagRequired("env-id")
}

func runDelete(cmd *cobra.Command, args []string) error {
	composeDir := filepath.Join("envs", flagDeleteEnvID)

	// 1. Verify the environment exists
	if _, err := os.Stat(composeDir); os.IsNotExist(err) {
		return fmt.Errorf("environment %q not found (looked in %s)", flagDeleteEnvID, composeDir)
	}

	fmt.Printf("🗑️  Deleting ephemeral environment %s...\n", flagDeleteEnvID)

	// 2. Capture container start time before bringing it down
	startedAt, err := docker.ContainerStartedAt(flagDeleteEnvID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "   ⚠️  Could not read container start time: %v\n", err)
	}
	deletedAt := time.Now()

	// 3. docker compose down --rmi all --volumes --remove-orphans
	fmt.Println("   → Stopping containers and removing images...")
	if err := docker.ComposeDown(composeDir); err != nil {
		fmt.Fprintf(os.Stderr, "   ⚠️  docker compose down failed: %v\n", err)
		fmt.Fprintln(os.Stderr, "   Continuing with local cleanup...")
	}

	// 4. Remove the envs/<env-id>/ directory
	fmt.Printf("   → Removing %s/...\n", composeDir)
	if err := os.RemoveAll(composeDir); err != nil {
		return fmt.Errorf("removing environment directory: %w", err)
	}

	fmt.Printf("\n✅ Environment %s deleted successfully.\n", flagDeleteEnvID)

	// 5. Print cost summary
	printCostSummary(startedAt, deletedAt)

	return nil
}

const costPerHour = 0.20

func printCostSummary(startedAt, deletedAt time.Time) {
	fmt.Println()
	if startedAt.IsZero() {
		fmt.Println("💰 Cost summary: container start time unavailable, cost could not be calculated.")
		return
	}

	duration := deletedAt.Sub(startedAt)
	hours := duration.Hours()
	cost := hours * costPerHour

	// Round up to the nearest minute for display
	totalMinutes := int(math.Ceil(duration.Minutes()))
	displayHours := totalMinutes / 60
	displayMins := totalMinutes % 60

	var durationStr string
	if displayHours > 0 {
		durationStr = fmt.Sprintf("%dh %02dm", displayHours, displayMins)
	} else {
		durationStr = fmt.Sprintf("%dm", displayMins)
	}

	fmt.Printf("💰 Cost summary:\n")
	fmt.Printf("   Uptime:  %s\n", durationStr)
	fmt.Printf("   Rate:    $%.2f/hour\n", costPerHour)
	fmt.Printf("   Total:   $%.4f\n", cost)
}
