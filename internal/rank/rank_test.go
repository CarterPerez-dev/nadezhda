// ©AngelaMos | 2026
// rank_test.go

package rank

import (
	"math"
	"testing"
	"time"

	"github.com/CarterPerez-dev/nadezhda/internal/config"
	"github.com/CarterPerez-dev/nadezhda/internal/store"
)

func f(v float64) *float64 { return &v }

func TestScoreGoldenOrder(t *testing.T) {
	cfg := config.Default().Rank

	a := Signals{
		AgeHours: 2, MaxCVSS: 9.8, KEV: true, MaxEPSS: 0.97,
		ClusterSize: 5, ClusterAgeHours: 6, SourceWeight: 1.0, KeywordMatch: true,
	}
	b := Signals{
		AgeHours: 120, MaxCVSS: 4.3, KEV: false, MaxEPSS: 0.02,
		ClusterSize: 1, ClusterAgeHours: 0, SourceWeight: 0.6, KeywordMatch: false,
	}

	sa, sb := Score(a, cfg), Score(b, cfg)
	if !(sa > sb) {
		t.Fatalf("expected A (%.4f) to rank strictly above B (%.4f)", sa, sb)
	}
	if sa < 0.95 {
		t.Errorf("A score = %.4f, want ~0.98", sa)
	}
	if sb > 0.20 {
		t.Errorf("B score = %.4f, want ~0.14", sb)
	}
}

func TestRecencyDecay(t *testing.T) {
	if got := recency(0, 48); math.Abs(got-1.0) > 1e-9 {
		t.Errorf("recency(0) = %v, want 1.0", got)
	}
	if got := recency(48, 48); math.Abs(got-0.5) > 1e-9 {
		t.Errorf("recency(one half-life) = %v, want 0.5", got)
	}
	if got := recency(-5, 48); got != 1.0 {
		t.Errorf("negative age should clamp to 1.0, got %v", got)
	}
}

func TestVelocity(t *testing.T) {
	if v := velocity(1, 0, 0.5); v != 0 {
		t.Errorf("size-1 velocity = %v, want 0", v)
	}
	if v := velocity(6, 12, 0.5); v != 1.0 {
		t.Errorf("size-6 in 12h should saturate, got %v", v)
	}
	if v := velocity(3, 0, 0.5); v != 1.0 {
		t.Errorf("burst (age<floor) should saturate, got %v", v)
	}
	if v := velocity(2, 100, 0.5); math.Abs(v-0.04) > 1e-9 {
		t.Errorf("slow cluster velocity = %v, want 0.04", v)
	}
}

func TestRankOrdersByScore(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	hot := store.DigestCluster{
		ClusterID: 1, Size: 4, FirstSeen: now.Unix() - 3600, LastSeen: now.Unix() - 60,
		Articles: []store.DigestArticle{{SourceName: "a", SourceWeight: 1.0, PublishedAt: now.Unix() - 60}},
		CVEs:     []store.DigestCVE{{ID: "CVE-2026-1", CVSSScore: f(9.9), EPSS: f(0.95), IsKEV: true}},
	}
	cold := store.DigestCluster{
		ClusterID: 2, Size: 1, FirstSeen: now.Unix() - 500*3600, LastSeen: now.Unix() - 500*3600,
		Articles: []store.DigestArticle{{SourceName: "b", SourceWeight: 0.5, PublishedAt: now.Unix() - 500*3600}},
	}

	scored := Rank([]store.DigestCluster{cold, hot}, config.Default().Rank, nil, now)
	if scored[0].Cluster.ClusterID != 1 {
		t.Errorf("hot KEV cluster should rank first, got cluster %d", scored[0].Cluster.ClusterID)
	}
	if !(scored[0].Score > scored[1].Score) {
		t.Errorf("scores not descending: %.4f then %.4f", scored[0].Score, scored[1].Score)
	}
}

func TestKeywordMatch(t *testing.T) {
	c := store.DigestCluster{
		Articles: []store.DigestArticle{{Title: "Fortinet FortiOS flaw exploited"}},
		CVEs:     []store.DigestCVE{{ID: "CVE-2026-1"}},
	}
	if !matchesWatchlist(c, []string{"fortinet"}) {
		t.Error("case-insensitive watchlist term should match title")
	}
	if !matchesWatchlist(c, []string{"CVE-2026-1"}) {
		t.Error("watchlist should match against CVE ids")
	}
	if matchesWatchlist(c, []string{"cisco"}) {
		t.Error("non-matching term should not match")
	}
	if matchesWatchlist(c, nil) {
		t.Error("empty watchlist should not match")
	}
}
