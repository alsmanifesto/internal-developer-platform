package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ravon/scaffold/internal/metadata"
	"github.com/ravon/scaffold/internal/tui"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an existing service via interactive TUI",
	Long:  `Launch an interactive TUI to select and delete an existing scaffolded service.`,
	RunE:  runDelete,
}

func runDelete(cmd *cobra.Command, args []string) error {
	store, err := metadata.LoadMetadata()
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	services := metadata.ListServices(store)
	if len(services) == 0 {
		fmt.Println("No services found.")
		return nil
	}

	m := tui.NewDeleteModel(services)
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	result, ok := finalModel.(tui.DeleteModel)
	if !ok {
		return fmt.Errorf("unexpected model type")
	}

	if !result.Confirmed() {
		fmt.Println("Aborted.")
		return nil
	}

	svcName := result.SelectedService()

	// 1. Run terraform destroy if terraform directory exists
	tfDir := filepath.Join(svcName, "terraform")
	if info, err := os.Stat(tfDir); err == nil && info.IsDir() {
		fmt.Printf("  → Running terraform destroy in %s...\n", tfDir)
		tfCmd := exec.Command("terraform", "destroy", "-auto-approve")
		tfCmd.Dir = tfDir
		tfCmd.Stdout = os.Stdout
		tfCmd.Stderr = os.Stderr
		if err := tfCmd.Run(); err != nil {
			fmt.Printf("  ⚠  terraform destroy failed (continuing): %v\n", err)
		} else {
			fmt.Println("  ✓ Infrastructure destroyed.")
		}
	}

	// 2. Delete the service folder
	svcPath := svcName
	if _, err := os.Stat(svcPath); err == nil {
		fmt.Printf("  → Deleting folder %s/...\n", svcPath)
		if err := os.RemoveAll(svcPath); err != nil {
			return fmt.Errorf("failed to delete service folder: %w", err)
		}
		fmt.Println("  ✓ Folder deleted.")
	} else {
		fmt.Printf("  ⚠  Folder %s not found, skipping.\n", svcPath)
	}

	// 3. Remove metadata entry and save
	if err := metadata.RemoveService(store, svcName); err != nil {
		return fmt.Errorf("failed to remove service from metadata: %w", err)
	}

	if err := metadata.SaveMetadata(store); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	fmt.Println("🧹 Service removed successfully")
	fmt.Printf("Repository %s deleted\n", svcName)

	return nil
}
