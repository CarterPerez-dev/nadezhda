// ©AngelaMos | 2026
// format.go

package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/CarterPerez-dev/nadezhda/internal/store"
)

type band int

const (
	bandNone band = iota
	bandLow
	bandMedium
	bandHigh
	bandCritical
)

const (
	cvssCritical = 9.0
	cvssHigh     = 7.0
	cvssMedium   = 4.0

	epssHot  = 0.5
	epssWarm = 0.1

	secsPerMinute = 60
	secsPerHour   = 3600
	secsPerDay    = 86400

	ellipsis    = "…"
	emptyMarker = "───"
	naMarker    = "—"
)

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func cvssBand(score *float64) band {
	if score == nil {
		return bandNone
	}
	switch v := *score; {
	case v >= cvssCritical:
		return bandCritical
	case v >= cvssHigh:
		return bandHigh
	case v >= cvssMedium:
		return bandMedium
	case v > 0:
		return bandLow
	default:
		return bandNone
	}
}

func bandLabel(b band) string {
	switch b {
	case bandCritical:
		return "CRITICAL"
	case bandHigh:
		return "HIGH"
	case bandMedium:
		return "MEDIUM"
	case bandLow:
		return "LOW"
	default:
		return naMarker
	}
}

func clusterHasKEV(c store.DigestCluster) bool {
	for _, v := range c.CVEs {
		if v.IsKEV {
			return true
		}
	}
	return false
}

func clusterMaxCVSS(c store.DigestCluster) *float64 {
	var best *float64
	for _, v := range c.CVEs {
		if v.CVSSScore == nil {
			continue
		}
		if best == nil || *v.CVSSScore > *best {
			best = v.CVSSScore
		}
	}
	return best
}

func clusterMaxEPSS(c store.DigestCluster) *float64 {
	var best *float64
	for _, v := range c.CVEs {
		if v.EPSS == nil {
			continue
		}
		if best == nil || *v.EPSS > *best {
			best = v.EPSS
		}
	}
	return best
}

func clusterBand(c store.DigestCluster) band {
	if clusterHasKEV(c) {
		return bandCritical
	}
	return cvssBand(clusterMaxCVSS(c))
}

func epssBand(p *float64) band {
	if p == nil {
		return bandNone
	}
	switch v := *p; {
	case v >= epssHot:
		return bandCritical
	case v >= epssWarm:
		return bandHigh
	case v > 0:
		return bandMedium
	default:
		return bandNone
	}
}

func cvssString(score *float64) string {
	if score == nil {
		return naMarker
	}
	return fmt.Sprintf("%.1f", *score)
}

func epssString(p *float64) string {
	if p == nil {
		return naMarker
	}
	return fmt.Sprintf("%.1f%%", clamp01(*p)*100)
}

func relativeAge(unix int64, now time.Time) string {
	if unix <= 0 {
		return naMarker
	}
	secs := now.Unix() - unix
	if secs < 0 {
		secs = 0
	}
	switch {
	case secs < secsPerMinute:
		return "just now"
	case secs < secsPerHour:
		return fmt.Sprintf("%dm ago", secs/secsPerMinute)
	case secs < secsPerDay:
		return fmt.Sprintf("%dh ago", secs/secsPerHour)
	default:
		return fmt.Sprintf("%dd ago", secs/secsPerDay)
	}
}

func truncate(s string, max int) string {
	s = strings.Join(strings.Fields(s), " ")
	r := []rune(s)
	if max <= 0 {
		return ""
	}
	if len(r) <= max {
		return s
	}
	if max == 1 {
		return ellipsis
	}
	return string(r[:max-1]) + ellipsis
}

func padRight(s string, width int) string {
	w := len([]rune(s))
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}
