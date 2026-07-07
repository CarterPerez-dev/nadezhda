// ©AngelaMos | 2026
// watch.go

package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/spf13/cobra"

	"github.com/CarterPerez-dev/nadezhda/internal/config"
	"github.com/CarterPerez-dev/nadezhda/internal/enrich"
	"github.com/CarterPerez-dev/nadezhda/internal/fetch"
	"github.com/CarterPerez-dev/nadezhda/internal/rank"
	"github.com/CarterPerez-dev/nadezhda/internal/source"
	"github.com/CarterPerez-dev/nadezhda/internal/store"
	"github.com/CarterPerez-dev/nadezhda/internal/watch"
)

const untitledCluster = "(untitled cluster)"

var (
	watchInterval string
	watchOnce     bool
	watchNoEnrich bool
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Run as a daemon, re-ingesting on an interval with an optional webhook notify",
	Long: "Run nadezhda as a long-lived daemon that re-ingests every enabled source on an interval " +
		"(default from config, override with --interval). Each cycle scrapes, clusters, and enriches " +
		"exactly like the scrape command. When watch.webhook_url is set, genuinely new high-signal " +
		"stories are POSTed to that webhook (Slack, Discord, or any JSON endpoint).",
	RunE: runWatch,
}

func init() {
	watchCmd.Flags().StringVar(&watchInterval, "interval", "", "re-ingest interval, e.g. 30m or 1h (overrides watch.interval in config)")
	watchCmd.Flags().BoolVar(&watchOnce, "once", false, "run a single cycle and exit instead of looping")
	watchCmd.Flags().BoolVar(&watchNoEnrich, "no-enrich", false, "skip the keyless CVE enrichment pass each cycle")
	rootCmd.AddCommand(watchCmd)
}

func runWatch(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	interval, err := resolveWatchInterval(cfg)
	if err != nil {
		return err
	}
	if cfg.Watch.WebhookURL != "" {
		if err := validateWebhookURL(cfg.Watch.WebhookURL); err != nil {
			return err
		}
	}

	srcs, err := source.Load(cfg.SourcesPath)
	if err != nil {
		return err
	}
	targets := source.Enabled(srcs)
	if len(targets) == 0 {
		return fmt.Errorf("watch: no enabled sources to poll")
	}

	st, err := store.Open(cfg.DBPath)
	if err != nil {
		return err
	}
	defer st.Close()

	fc := fetch.New(fetch.Options{
		UserAgent:    cfg.Fetch.UserAgent,
		PerHostRate:  cfg.Fetch.PerHostRate,
		PerHostBurst: cfg.Fetch.PerHostBurst,
		Timeout:      time.Duration(cfg.Fetch.TimeoutSeconds) * time.Second,
		MaxRetries:   cfg.Fetch.MaxRetries,
	})

	out := cmd.ErrOrStderr()
	if watchNoEnrich && cfg.Watch.NotifyOnKEV {
		fmt.Fprintln(out, "watch: --no-enrich with notify_on_kev set: KEV status is never computed, so KEV alerts will not fire")
	}

	cycle := func(ctx context.Context) (watch.Report, error) {
		start := time.Now()
		summary, cstats, err := ingestAndCluster(ctx, fc, st, cfg, targets, start)
		if err != nil {
			return watch.Report{}, err
		}
		newArticles, duplicates, failed := summary.Totals()

		var enriched, kevHits int
		if !watchNoEnrich {
			enriched, kevHits = watchEnrich(ctx, out, st, cfg)
		}

		notable, err := buildNotable(st, cfg, start)
		if err != nil {
			return watch.Report{}, err
		}
		return watch.Report{
			Start:       start,
			Duration:    time.Since(start),
			NewArticles: newArticles,
			Duplicates:  duplicates,
			Clusters:    cstats.Total,
			Enriched:    enriched,
			KEVHits:     kevHits,
			Failed:      failed,
			Notable:     notable,
		}, nil
	}

	opts := watch.Options{
		Interval:   interval,
		RunAtStart: true,
		Cycle:      cycle,
		Out:        out,
	}
	if cfg.Watch.WebhookURL != "" {
		opts.Notifier = watch.WebhookNotifier{
			URL:    cfg.Watch.WebhookURL,
			Client: &http.Client{Timeout: time.Duration(cfg.Fetch.TimeoutSeconds) * time.Second},
		}
	}

	if watchOnce {
		return watch.Once(cmd.Context(), opts)
	}
	return watch.Run(cmd.Context(), opts)
}

func resolveWatchInterval(cfg config.Config) (time.Duration, error) {
	raw := cfg.Watch.Interval
	if watchInterval != "" {
		raw = watchInterval
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("watch: invalid interval %q: %w", raw, err)
	}
	if d < config.MinWatchInterval {
		return 0, fmt.Errorf("watch: interval %s is below the minimum %s", d, config.MinWatchInterval)
	}
	return d, nil
}

func validateWebhookURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("watch: invalid webhook_url %q: %w", raw, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("watch: webhook_url must be http or https, got %q", raw)
	}
	if u.Host == "" {
		return fmt.Errorf("watch: webhook_url must include a host, got %q", raw)
	}
	return nil
}

func watchEnrich(ctx context.Context, out io.Writer, st *store.Store, cfg config.Config) (enriched, kevHits int) {
	ectx, cancel := context.WithTimeout(ctx, enrichBudget)
	defer cancel()
	stats, err := enrich.Run(ectx, st, buildEnrichClients(cfg), time.Now(), cfg.Enrich.CacheTTLHours, cfg.Enrich.NegativeTTLHours)
	if err != nil && ctx.Err() == nil {
		fmt.Fprintf(out, "watch: enrich degraded, KEV/CVE data may be stale: %v\n", err)
	}
	return stats.Enriched, stats.KEVHits
}

func buildNotable(st *store.Store, cfg config.Config, cycleStart time.Time) ([]watch.NotableItem, error) {
	fresh, err := st.NewlyFetchedClusters(cycleStart.Unix())
	if err != nil {
		return nil, err
	}
	scored := rank.Rank(fresh, cfg.Rank, cfg.Watchlist, time.Now())
	items := make([]watch.NotableItem, 0, cfg.Watch.NotifyMaxItems)
	for _, sc := range scored {
		if len(items) >= cfg.Watch.NotifyMaxItems {
			break
		}
		if isNotable(sc, cfg.Watch) {
			items = append(items, toNotable(sc))
		}
	}
	return items, nil
}

func isNotable(sc rank.Scored, w config.Watch) bool {
	if sc.Score >= w.NotifyMinScore {
		return true
	}
	if w.NotifyOnKEV {
		for _, v := range sc.Cluster.CVEs {
			if v.IsKEV {
				return true
			}
		}
	}
	return false
}

func toNotable(sc rank.Scored) watch.NotableItem {
	c := sc.Cluster
	title, link := representativeArticle(c.Articles)
	var maxCVSS float64
	var isKEV bool
	cves := make([]string, 0, len(c.CVEs))
	for _, v := range c.CVEs {
		cves = append(cves, v.ID)
		if v.CVSSScore != nil && *v.CVSSScore > maxCVSS {
			maxCVSS = *v.CVSSScore
		}
		if v.IsKEV {
			isKEV = true
		}
	}
	return watch.NotableItem{
		Title:   title,
		URL:     link,
		Score:   sc.Score,
		MaxCVSS: maxCVSS,
		IsKEV:   isKEV,
		CVEs:    cves,
		Sources: distinctSources(c.Articles),
	}
}

func representativeArticle(articles []store.DigestArticle) (title, link string) {
	if len(articles) == 0 {
		return untitledCluster, ""
	}
	best := articles[0]
	for _, a := range articles[1:] {
		if a.SourceWeight > best.SourceWeight {
			best = a
		}
	}
	return best.Title, best.CanonicalURL
}

func distinctSources(articles []store.DigestArticle) int {
	set := make(map[string]struct{}, len(articles))
	for _, a := range articles {
		set[a.SourceName] = struct{}{}
	}
	return len(set)
}
