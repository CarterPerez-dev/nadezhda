// ©AngelaMos | 2026
// watch_test.go

package main

import (
	"fmt"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/CarterPerez-dev/nadezhda/internal/config"
	"github.com/CarterPerez-dev/nadezhda/internal/rank"
	"github.com/CarterPerez-dev/nadezhda/internal/store"
)

func TestIsNotable(t *testing.T) {
	kevCluster := store.DigestCluster{CVEs: []store.DigestCVE{{ID: "CVE-2026-1", IsKEV: true}}}
	plainCluster := store.DigestCluster{}
	tests := []struct {
		name   string
		scored rank.Scored
		watch  config.Watch
		want   bool
	}{
		{"score above threshold", rank.Scored{Cluster: plainCluster, Score: 0.9}, config.Watch{NotifyMinScore: 0.5}, true},
		{"score below threshold, not kev", rank.Scored{Cluster: plainCluster, Score: 0.2}, config.Watch{NotifyMinScore: 0.5}, false},
		{"below threshold but kev with notify_on_kev", rank.Scored{Cluster: kevCluster, Score: 0.1}, config.Watch{NotifyMinScore: 0.5, NotifyOnKEV: true}, true},
		{"kev but notify_on_kev off", rank.Scored{Cluster: kevCluster, Score: 0.1}, config.Watch{NotifyMinScore: 0.5, NotifyOnKEV: false}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNotable(tt.scored, tt.watch); got != tt.want {
				t.Errorf("isNotable = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToNotablePicksHighestAuthoritySource(t *testing.T) {
	cvss := 9.8
	sc := rank.Scored{
		Score: 0.77,
		Cluster: store.DigestCluster{
			Articles: []store.DigestArticle{
				{Title: "low authority", CanonicalURL: "https://low", SourceName: "low", SourceWeight: 0.5},
				{Title: "high authority", CanonicalURL: "https://high", SourceName: "high", SourceWeight: 1.0},
			},
			CVEs: []store.DigestCVE{{ID: "CVE-2026-1", CVSSScore: &cvss, IsKEV: true}},
		},
	}
	n := toNotable(sc)
	if n.Title != "high authority" || n.URL != "https://high" {
		t.Errorf("representative should be the highest-source-weight article, got %q / %q", n.Title, n.URL)
	}
	if !n.IsKEV || n.MaxCVSS != 9.8 {
		t.Errorf("cve signals wrong: kev=%v cvss=%v", n.IsKEV, n.MaxCVSS)
	}
	if n.Sources != 2 {
		t.Errorf("distinct sources = %d, want 2", n.Sources)
	}
	if len(n.CVEs) != 1 || n.CVEs[0] != "CVE-2026-1" {
		t.Errorf("cves wrong: %v", n.CVEs)
	}
	if n.Score != 0.77 {
		t.Errorf("score = %v, want 0.77", n.Score)
	}
}

func TestResolveWatchInterval(t *testing.T) {
	cfg := config.Default()
	watchInterval = ""
	defer func() { watchInterval = "" }()

	if d, err := resolveWatchInterval(cfg); err != nil || d != time.Hour {
		t.Errorf("default interval = %v, %v; want 1h, nil", d, err)
	}

	watchInterval = "30m"
	if d, err := resolveWatchInterval(cfg); err != nil || d != 30*time.Minute {
		t.Errorf("flag override = %v, %v; want 30m, nil", d, err)
	}

	watchInterval = "5s"
	if _, err := resolveWatchInterval(cfg); err == nil {
		t.Error("expected an error for a sub-minimum interval")
	}

	watchInterval = "nonsense"
	if _, err := resolveWatchInterval(cfg); err == nil {
		t.Error("expected an error for a garbage interval")
	}
}

func TestValidateWebhookURL(t *testing.T) {
	good := []string{"https://hooks.slack.com/services/x", "http://127.0.0.1:39871/hook"}
	for _, u := range good {
		if err := validateWebhookURL(u); err != nil {
			t.Errorf("validateWebhookURL(%q) = %v, want nil", u, err)
		}
	}
	bad := []string{"example.com/hook", "ftp://x/y", "not a url", ""}
	for _, u := range bad {
		if err := validateWebhookURL(u); err == nil {
			t.Errorf("validateWebhookURL(%q) = nil, want error", u)
		}
	}
}

func TestBuildNotableCapsToNotifyMax(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "watch.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	src, err := st.UpsertSource(store.SourceInput{Name: "s", URL: "https://s/f", Type: "rss", Weight: 1, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	var rows []store.ClusterRow
	for i := 0; i < 5; i++ {
		id, err := st.InsertArticle(store.Article{
			SourceID: src, CanonicalURL: fmt.Sprintf("https://s/%d", i),
			ContentHash: "c" + strconv.Itoa(i), TitleHash: "t" + strconv.Itoa(i),
			Title: "story " + strconv.Itoa(i), PublishedAt: 10000, FetchedAt: 10000,
		})
		if err != nil {
			t.Fatal(err)
		}
		rows = append(rows, store.ClusterRow{Key: strconv.Itoa(i), Members: []int64{id}, FirstSeen: 10000, LastSeen: 10000})
	}
	if err := st.ReplaceClusters(rows); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	cfg.Watch.NotifyMinScore = 0
	cfg.Watch.NotifyMaxItems = 2

	items, err := buildNotable(st, cfg, time.Unix(9999, 0))
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("buildNotable returned %d items, want the notify_max_items cap of 2", len(items))
	}
}
