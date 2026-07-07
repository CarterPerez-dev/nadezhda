// ©AngelaMos | 2026
// digest.go

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/CarterPerez-dev/nadezhda/internal/export"
	"github.com/CarterPerez-dev/nadezhda/internal/rank"
	"github.com/CarterPerez-dev/nadezhda/internal/store"
)

const (
	defaultDigestTop = 20
	formatMarkdown   = "md"
	formatJSON       = "json"
	outFilePerm      = 0o644
)

var (
	digestTop    int
	digestSince  string
	digestFormat string
	digestOut    string
)

var digestCmd = &cobra.Command{
	Use:   "digest",
	Short: "Render a ranked digest of story clusters to Markdown or JSON",
	RunE:  runDigest,
}

func init() {
	digestCmd.Flags().IntVar(&digestTop, "top", defaultDigestTop, "show the top N ranked clusters")
	digestCmd.Flags().StringVar(&digestSince, "since", "", "only clusters active within this window (e.g. 24h, 168h)")
	digestCmd.Flags().StringVar(&digestFormat, "format", formatMarkdown, "output format: md or json")
	digestCmd.Flags().StringVar(&digestOut, "out", "", "write to this file instead of stdout")
	rootCmd.AddCommand(digestCmd)
}

func runDigest(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	if digestFormat != formatMarkdown && digestFormat != formatJSON {
		return fmt.Errorf("invalid --format %q: want md or json", digestFormat)
	}

	var since int64
	now := time.Now()
	if digestSince != "" {
		d, err := time.ParseDuration(digestSince)
		if err != nil {
			return fmt.Errorf("invalid --since %q: %w", digestSince, err)
		}
		since = now.Add(-d).Unix()
	}

	st, err := store.Open(cfg.DBPath)
	if err != nil {
		return err
	}
	defer st.Close()

	clusters, err := st.DigestClusters(since)
	if err != nil {
		return err
	}
	scored := rank.Rank(clusters, cfg.Rank, cfg.Watchlist, now)

	shown := len(scored)
	if digestTop > 0 && digestTop < shown {
		shown = digestTop
	}

	var rendered string
	if digestFormat == formatJSON {
		rendered, err = export.JSON(scored, digestTop)
		if err != nil {
			return err
		}
	} else {
		rendered = export.Markdown(scored, digestTop)
	}

	if digestOut != "" {
		if err := os.WriteFile(digestOut, []byte(rendered), outFilePerm); err != nil {
			return fmt.Errorf("write digest to %s: %w", digestOut, err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "wrote %d clusters to %s\n", shown, digestOut)
		return nil
	}
	fmt.Fprint(cmd.OutOrStdout(), rendered)
	return nil
}
