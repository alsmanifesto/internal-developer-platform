package cmd

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ravon/scaffold/internal/generator"
	"github.com/ravon/scaffold/internal/metadata"
	"github.com/ravon/scaffold/internal/tui"
	"github.com/spf13/cobra"
)

var (
	flagName        string
	flagServiceType string
	flagWorkload    string
	flagStack       string
	flagPipeline    string
)

var validServiceTypes = []string{"api", "worker", "job"}
var validWorkloads = []string{"app", "data", "ml"}
var validStacks = []string{"go", "python", "spark", "kafka"}
var validPipelines = []string{"gh-actions", "concourse", "airflow", "mlflow"}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new service (interactive TUI or --flags)",
	Long: `Scaffold a new service interactively or non-interactively.

Interactive (default):
  scaffold create

Non-interactive (all flags required):
  scaffold create --name my-svc --service-type api --workload app --stack go --pipeline gh-actions`,
	RunE: runCreate,
}

func init() {
	createCmd.Flags().StringVar(&flagName, "name", "", "Service name")
	createCmd.Flags().StringVar(&flagServiceType, "service-type", "", "Service type: api | worker | job")
	createCmd.Flags().StringVar(&flagWorkload, "workload", "", "Workload: app | data | ml")
	createCmd.Flags().StringVar(&flagStack, "stack", "", "Stack: go | python | spark | kafka")
	createCmd.Flags().StringVar(&flagPipeline, "pipeline", "", "Pipeline: gh-actions | concourse | airflow | mlflow")
}

func runCreate(cmd *cobra.Command, args []string) error {
	// If any flag is provided, run in non-interactive mode
	if flagName != "" || flagServiceType != "" || flagWorkload != "" || flagStack != "" || flagPipeline != "" {
		return runCreateCLI()
	}
	return runCreateTUI()
}

func runCreateCLI() error {
	var errs []string

	if strings.TrimSpace(flagName) == "" {
		errs = append(errs, "  --name is required")
	}
	if !contains(validServiceTypes, flagServiceType) {
		errs = append(errs, fmt.Sprintf("  --service-type must be one of: %s", strings.Join(validServiceTypes, ", ")))
	}
	if !contains(validWorkloads, flagWorkload) {
		errs = append(errs, fmt.Sprintf("  --workload must be one of: %s", strings.Join(validWorkloads, ", ")))
	}
	if !contains(validStacks, flagStack) {
		errs = append(errs, fmt.Sprintf("  --stack must be one of: %s", strings.Join(validStacks, ", ")))
	}
	if !contains(validPipelines, flagPipeline) {
		errs = append(errs, fmt.Sprintf("  --pipeline must be one of: %s", strings.Join(validPipelines, ", ")))
	}

	if len(errs) > 0 {
		return fmt.Errorf("invalid flags:\n%s", strings.Join(errs, "\n"))
	}

	svc := metadata.ServiceMetadata{
		Name:        strings.TrimSpace(flagName),
		ServiceType: flagServiceType,
		Workload:    flagWorkload,
		Stack:       flagStack,
		Pipeline:    flagPipeline,
		Path:        strings.TrimSpace(flagName),
		CreatedAt:   time.Now(),
	}

	return scaffold(svc)
}

func runCreateTUI() error {
	m := tui.NewCreateModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	result, ok := finalModel.(tui.CreateModel)
	if !ok {
		return fmt.Errorf("unexpected model type")
	}

	if !result.Confirmed() {
		fmt.Println("Aborted.")
		return nil
	}

	return scaffold(result.ServiceConfig())
}

func scaffold(svc metadata.ServiceMetadata) error {
	fmt.Println("⚡ Ravon is creating your service...")

	store, err := metadata.LoadMetadata()
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	if err := metadata.AddService(store, svc); err != nil {
		return fmt.Errorf("failed to add service to metadata: %w", err)
	}

	if err := generator.Generate(svc); err != nil {
		return fmt.Errorf("failed to generate scaffolding: %w", err)
	}

	if err := metadata.SaveMetadata(store); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	fmt.Println("✅ Service successfully created by Ravon")
	fmt.Printf("   Service path: %s/\n", svc.Name)
	return nil
}

func contains(list []string, val string) bool {
	for _, v := range list {
		if v == val {
			return true
		}
	}
	return false
}
