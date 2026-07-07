// ©AngelaMos | 2026
// export.go

package export

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/CarterPerez-dev/nadezhda/internal/rank"
	"github.com/CarterPerez-dev/nadezhda/internal/store"
)

const (
	dateLayout   = "2006-01-02"
	noDate       = "----------"
	scoreRounder = 10000.0
	digestTitle  = "# Nadezhda Digest"
)

type CVEEntry struct {
	ID   string   `json:"id"`
	CVSS *float64 `json:"cvss,omitempty"`
	EPSS *float64 `json:"epss,omitempty"`
	KEV  bool     `json:"kev"`
}

type Entry struct {
	Rank      int        `json:"rank"`
	Score     float64    `json:"score"`
	Headline  string     `json:"headline"`
	URL       string     `json:"url"`
	Outlets   []string   `json:"outlets"`
	FirstSeen string     `json:"first_seen"`
	LastSeen  string     `json:"last_seen"`
	CVEs      []CVEEntry `json:"cves,omitempty"`
}

type Digest struct {
	Count   int     `json:"count"`
	Entries []Entry `json:"entries"`
}

func Build(scored []rank.Scored, top int) Digest {
	limit := len(scored)
	if top > 0 && top < limit {
		limit = top
	}
	entries := make([]Entry, 0, limit)
	for i := 0; i < limit; i++ {
		s := scored[i]
		h := headline(s.Cluster.Articles)
		entries = append(entries, Entry{
			Rank:      i + 1,
			Score:     roundScore(s.Score),
			Headline:  h.Title,
			URL:       h.CanonicalURL,
			Outlets:   outlets(s.Cluster.Articles),
			FirstSeen: formatDate(s.Cluster.FirstSeen),
			LastSeen:  formatDate(s.Cluster.LastSeen),
			CVEs:      cveEntries(s.Cluster.CVEs),
		})
	}
	return Digest{Count: len(entries), Entries: entries}
}

func JSON(scored []rank.Scored, top int) (string, error) {
	b, err := json.MarshalIndent(Build(scored, top), "", "  ")
	if err != nil {
		return "", fmt.Errorf("export json: %w", err)
	}
	return string(b), nil
}

func Markdown(scored []rank.Scored, top int) string {
	d := Build(scored, top)
	var b strings.Builder
	b.WriteString(digestTitle + "\n\n")
	for _, e := range d.Entries {
		fmt.Fprintf(&b, "## %d. %s\n", e.Rank, oneLine(e.Headline))
		fmt.Fprintf(&b, "score %.2f | %d outlet(s): %s\n", e.Score, len(e.Outlets), strings.Join(e.Outlets, ", "))
		for _, c := range e.CVEs {
			fmt.Fprintf(&b, "- %s\n", cveLine(c))
		}
		fmt.Fprintf(&b, "%s\n\n", e.URL)
	}
	return b.String()
}

func headline(articles []store.DigestArticle) store.DigestArticle {
	var best store.DigestArticle
	for i, a := range articles {
		if i == 0 || a.PublishedAt > best.PublishedAt ||
			(a.PublishedAt == best.PublishedAt && a.ID < best.ID) {
			best = a
		}
	}
	return best
}

func outlets(articles []store.DigestArticle) []string {
	seen := make(map[string]struct{}, len(articles))
	var out []string
	for _, a := range articles {
		if _, ok := seen[a.SourceName]; ok {
			continue
		}
		seen[a.SourceName] = struct{}{}
		out = append(out, a.SourceName)
	}
	sort.Strings(out)
	return out
}

func cveEntries(cves []store.DigestCVE) []CVEEntry {
	out := make([]CVEEntry, 0, len(cves))
	for _, v := range cves {
		out = append(out, CVEEntry{ID: v.ID, CVSS: v.CVSSScore, EPSS: v.EPSS, KEV: v.IsKEV})
	}
	sort.SliceStable(out, func(a, b int) bool {
		if out[a].KEV != out[b].KEV {
			return out[a].KEV
		}
		if cvssOf(out[a]) != cvssOf(out[b]) {
			return cvssOf(out[a]) > cvssOf(out[b])
		}
		return out[a].ID < out[b].ID
	})
	return out
}

func cvssOf(c CVEEntry) float64 {
	if c.CVSS == nil {
		return -1
	}
	return *c.CVSS
}

func cveLine(c CVEEntry) string {
	var parts []string
	if c.KEV {
		parts = append(parts, "KEV")
	}
	if c.CVSS != nil {
		parts = append(parts, fmt.Sprintf("CVSS %.1f", *c.CVSS))
	}
	if c.EPSS != nil {
		parts = append(parts, fmt.Sprintf("EPSS %.2f", *c.EPSS))
	}
	if len(parts) == 0 {
		return c.ID
	}
	return fmt.Sprintf("%s (%s)", c.ID, strings.Join(parts, ", "))
}

func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func roundScore(v float64) float64 {
	return float64(int64(v*scoreRounder+0.5)) / scoreRounder
}

func formatDate(unix int64) string {
	if unix == 0 {
		return noDate
	}
	return time.Unix(unix, 0).UTC().Format(dateLayout)
}
