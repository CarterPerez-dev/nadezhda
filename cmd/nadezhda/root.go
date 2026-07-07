// ©AngelaMos | 2026
// root.go

package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/CarterPerez-dev/nadezhda/internal/config"
	"github.com/CarterPerez-dev/nadezhda/internal/setup"
)

var (
	flagConfig string
	flagDB     string
)

var rootCmd = &cobra.Command{
	Use:          "nadezhda",
	Short:        "Security news and CVE aggregation engine",
	Long:         "Nadezhda aggregates cybersecurity news, enriches CVEs with NVD, CISA KEV, and EPSS intelligence, clusters stories across outlets, and ranks what matters.",
	SilenceUsage: true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagConfig, "config", "", "path to config.yaml (optional)")
	rootCmd.PersistentFlags().StringVar(&flagDB, "db", "", "override database path")
}

func loadConfig() (config.Config, error) {
	if err := setup.Load(); err != nil {
		return config.Config{}, err
	}
	cfg, err := config.Load(flagConfig)
	if err != nil {
		return config.Config{}, err
	}
	if flagDB != "" {
		cfg.DBPath = flagDB
	}
	if p := os.Getenv(setup.EnvProvider); p != "" {
		cfg.AI.Enabled = true
		cfg.AI.Provider = p
	}
	if u := os.Getenv(setup.EnvQwenURL); u != "" {
		cfg.AI.Qwen.BaseURL = u
	}
	return cfg, nil
}

func isInteractive(cmd *cobra.Command) bool {
	f, ok := cmd.InOrStdin().(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}
