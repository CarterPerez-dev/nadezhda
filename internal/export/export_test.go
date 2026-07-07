// ©AngelaMos | 2026
// export_test.go

package export

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/CarterPerez-dev/nadezhda/internal/rank"
	"github.com/CarterPerez-dev/nadezhda/internal/store"
)

func f(v float64) *float64 { return &v }

func sample() []rank.Scored {
	return []rank.Scored{
		{
			Score: 0.9412,
			Cluster: store.DigestCluster{
				Size: 3, FirstSeen: 1751600000, LastSeen: 1751700000,
				Articles: []store.DigestArticle{
					{ID: 1, SourceName: "thehackernews", Title: "Old framing", CanonicalURL: "https://thn/1", PublishedAt: 100},
					{ID: 2, SourceName: "bleepingcomputer", Title: "Freshest framing", CanonicalURL: "https://bc/2", PublishedAt: 300},
					{ID: 3, SourceName: "bleepingcomputer", Title: "Mid framing", CanonicalURL: "https://bc/3", PublishedAt: 200},
				},
				CVEs: []store.DigestCVE{
					{ID: "CVE-2026-2", CVSSScore: f(7.5), IsKEV: false},
					{ID: "CVE-2026-1", CVSSScore: f(9.8), EPSS: f(0.91), IsKEV: true},
				},
			},
		},
		{
			Score: 0.10,
			Cluster: store.DigestCluster{
				Size: 1, FirstSeen: 1751000000, LastSeen: 1751000000,
				Articles: []store.DigestArticle{{ID: 9, SourceName: "krebs", Title: "Quiet story", CanonicalURL: "https://k/9", PublishedAt: 50}},
			},
		},
	}
}

func TestBuildHeadlineAndOutlets(t *testing.T) {
	d := Build(sample(), 0)
	if d.Count != 2 {
		t.Fatalf("count = %d, want 2", d.Count)
	}
	e := d.Entries[0]
	if e.Rank != 1 || e.Score != 0.9412 {
		t.Errorf("rank/score = %d/%v", e.Rank, e.Score)
	}
	if e.Headline != "Freshest framing" {
		t.Errorf("headline = %q, want the freshest member", e.Headline)
	}
	if e.URL != "https://bc/2" {
		t.Errorf("url = %q, want the freshest member's", e.URL)
	}
	if len(e.Outlets) != 2 || e.Outlets[0] != "bleepingcomputer" || e.Outlets[1] != "thehackernews" {
		t.Errorf("outlets = %v, want distinct + sorted", e.Outlets)
	}
	if e.CVEs[0].ID != "CVE-2026-1" {
		t.Errorf("first cve = %q, want the KEV one first", e.CVEs[0].ID)
	}
}

func TestBuildTopLimit(t *testing.T) {
	d := Build(sample(), 1)
	if d.Count != 1 {
		t.Errorf("top=1 should yield 1 entry, got %d", d.Count)
	}
}

func TestJSONValidAndStable(t *testing.T) {
	out, err := JSON(sample(), 0)
	if err != nil {
		t.Fatal(err)
	}
	var back Digest
	if err := json.Unmarshal([]byte(out), &back); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if back.Entries[0].Headline != "Freshest framing" {
		t.Errorf("round-trip headline = %q", back.Entries[0].Headline)
	}
	if !back.Entries[0].CVEs[0].KEV {
		t.Error("first cve should be the KEV one")
	}
}

func TestMarkdownContainsSignals(t *testing.T) {
	md := Markdown(sample(), 0)
	for _, want := range []string{
		"# Nadezhda Digest",
		"## 1. Freshest framing",
		"2 outlet(s): bleepingcomputer, thehackernews",
		"CVE-2026-1 (KEV, CVSS 9.8, EPSS 0.91)",
		"https://bc/2",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("markdown missing %q\n---\n%s", want, md)
		}
	}
}
