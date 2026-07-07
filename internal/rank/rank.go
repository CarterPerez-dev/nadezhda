// ©AngelaMos | 2026
// rank.go

package rank

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/CarterPerez-dev/nadezhda/internal/config"
	"github.com/CarterPerez-dev/nadezhda/internal/store"
)

const (
	cvssMax                = 10.0
	minVelocityWindowHours = 1.0
	secondsPerHour         = 3600.0
)

type Signals struct {
	AgeHours        float64
	MaxCVSS         float64
	KEV             bool
	MaxEPSS         float64
	ClusterSize     int
	ClusterAgeHours float64
	SourceWeight    float64
	KeywordMatch    bool
}

func Score(s Signals, cfg config.Rank) float64 {
	w := cfg.Weights
	return w.Recency*recency(s.AgeHours, cfg.HalfLifeHours) +
		w.CVSS*clamp01(s.MaxCVSS/cvssMax) +
		w.KEV*boolScore(s.KEV) +
		w.EPSS*clamp01(s.MaxEPSS) +
		w.Velocity*velocity(s.ClusterSize, s.ClusterAgeHours, cfg.VelocityNorm) +
		w.Source*clamp01(s.SourceWeight) +
		w.Keyword*boolScore(s.KeywordMatch)
}

func recency(ageHours float64, halfLifeHours int) float64 {
	if ageHours < 0 {
		ageHours = 0
	}
	if halfLifeHours < 1 {
		halfLifeHours = 1
	}
	return math.Exp(-math.Ln2 * ageHours / float64(halfLifeHours))
}

func velocity(size int, ageHours, norm float64) float64 {
	if size <= 1 || norm <= 0 {
		return 0
	}
	if ageHours < minVelocityWindowHours {
		ageHours = minVelocityWindowHours
	}
	return clamp01((float64(size) / ageHours) / norm)
}

func boolScore(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

type Scored struct {
	Cluster store.DigestCluster
	Score   float64
}

func Rank(clusters []store.DigestCluster, cfg config.Rank, watchlist []string, now time.Time) []Scored {
	out := make([]Scored, len(clusters))
	for i, c := range clusters {
		out[i] = Scored{Cluster: c, Score: Score(signalsFor(c, watchlist, now), cfg)}
	}
	sort.SliceStable(out, func(a, b int) bool {
		if out[a].Score != out[b].Score {
			return out[a].Score > out[b].Score
		}
		return out[a].Cluster.LastSeen > out[b].Cluster.LastSeen
	})
	return out
}

func signalsFor(c store.DigestCluster, watchlist []string, now time.Time) Signals {
	s := Signals{
		AgeHours:        hoursSince(c.LastSeen, now),
		ClusterSize:     c.Size,
		ClusterAgeHours: float64(c.LastSeen-c.FirstSeen) / secondsPerHour,
		SourceWeight:    maxSourceWeight(c.Articles),
		KeywordMatch:    matchesWatchlist(c, watchlist),
	}
	for _, v := range c.CVEs {
		if v.CVSSScore != nil && *v.CVSSScore > s.MaxCVSS {
			s.MaxCVSS = *v.CVSSScore
		}
		if v.EPSS != nil && *v.EPSS > s.MaxEPSS {
			s.MaxEPSS = *v.EPSS
		}
		if v.IsKEV {
			s.KEV = true
		}
	}
	return s
}

func hoursSince(unix int64, now time.Time) float64 {
	return float64(now.Unix()-unix) / secondsPerHour
}

func maxSourceWeight(articles []store.DigestArticle) float64 {
	var max float64
	for _, a := range articles {
		if a.SourceWeight > max {
			max = a.SourceWeight
		}
	}
	return max
}

func matchesWatchlist(c store.DigestCluster, watchlist []string) bool {
	if len(watchlist) == 0 {
		return false
	}
	var sb strings.Builder
	for _, a := range c.Articles {
		sb.WriteString(strings.ToLower(a.Title))
		sb.WriteByte(' ')
	}
	for _, v := range c.CVEs {
		sb.WriteString(strings.ToLower(v.ID))
		sb.WriteByte(' ')
	}
	hay := sb.String()
	for _, term := range watchlist {
		if term == "" {
			continue
		}
		if strings.Contains(hay, strings.ToLower(term)) {
			return true
		}
	}
	return false
}
