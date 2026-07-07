// ©AngelaMos | 2026
// ai.go

package main

import (
	"github.com/spf13/cobra"

	"github.com/CarterPerez-dev/nadezhda/internal/setup"
)

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "Set up an AI provider for ideation (interactive, re-runnable)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return setup.Run(cmd.InOrStdin(), cmd.OutOrStdout())
	},
}

func init() {
	rootCmd.AddCommand(aiCmd)
}
