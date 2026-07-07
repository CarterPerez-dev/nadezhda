// ©AngelaMos | 2026
// sources.go

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/CarterPerez-dev/nadezhda/internal/source"
	"github.com/CarterPerez-dev/nadezhda/internal/store"
)

var sourcesCmd = &cobra.Command{
	Use:   "sources",
	Short: "List configured sources and persist them to the store",
	RunE:  runSources,
}

func init() {
	rootCmd.AddCommand(sourcesCmd)
}

func runSources(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	srcs, err := source.Load(cfg.SourcesPath)
	if err != nil {
		return err
	}
	st, err := store.Open(cfg.DBPath)
	if err != nil {
		return err
	}
	defer st.Close()

	for _, s := range srcs {
		if _, err := st.UpsertSource(store.SourceInput{
			Name: s.Name, Title: s.Title, URL: s.URL, Type: string(s.Type),
			Weight: s.Weight, Tags: s.Tags, Enabled: s.Enabled,
		}); err != nil {
			return err
		}
	}

	fmt.Printf("%-18s %-8s %-7s %s\n", "NAME", "ENABLED", "WEIGHT", "URL")
	for _, s := range srcs {
		enabled := "no"
		if s.Enabled {
			enabled = "yes"
		}
		fmt.Printf("%-18s %-8s %-7.2f %s\n", s.Name, enabled, s.Weight, s.URL)
	}
	fmt.Printf("\n%d sources (%d enabled) persisted to %s\n",
		len(srcs), len(source.Enabled(srcs)), cfg.DBPath)
	return nil
}
