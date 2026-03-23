package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ephemeral-env",
	Short: "Ephemeral environment manager for ravon services",
	Long: `ephemeral-env spins up preview environments for services scaffolded by ravon.
It generates a docker-compose configuration with Traefik routing and brings
the stack up via the Docker daemon.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(createCmd)
}
