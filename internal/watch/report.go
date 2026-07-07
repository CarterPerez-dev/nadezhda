// ©AngelaMos | 2026
// report.go

package watch

import "time"

type NotableItem struct {
	Title   string
	URL     string
	Score   float64
	MaxCVSS float64
	IsKEV   bool
	CVEs    []string
	Sources int
}

type Report struct {
	Start       time.Time
	Duration    time.Duration
	NewArticles int
	Duplicates  int
	Clusters    int
	Enriched    int
	KEVHits     int
	Failed      int
	Notable     []NotableItem
}
