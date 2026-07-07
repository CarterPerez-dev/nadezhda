// ©AngelaMos | 2026
// version.go

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/CarterPerez-dev/nadezhda/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%s %s\n", version.Name, version.Version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
