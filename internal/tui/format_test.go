// ©AngelaMos | 2026
// format_test.go

package tui

import (
	"testing"
	"time"

	"github.com/CarterPerez-dev/nadezhda/internal/store"
)

func ptr(v float64) *float64 { return &v }

func TestCVSSBand(t *testing.T) {
	cases := []struct {
		in   *float64
		want band
	}{
		{nil, bandNone},
		{ptr(0), bandNone},
		{ptr(3.9), bandLow},
		{ptr(4.0), bandMedium},
		{ptr(6.9), bandMedium},
		{ptr(7.0), bandHigh},
		{ptr(8.9), bandHigh},
		{ptr(9.0), bandCritical},
		{ptr(10.0), bandCritical},
	}
	for _, c := range cases {
		if got := cvssBand(c.in); got != c.want {
			t.Errorf("cvssBand(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestEPSSBand(t *testing.T) {
	cases := []struct {
		in   *float64
		want band
	}{
		{nil, bandNone},
		{ptr(0), bandNone},
		{ptr(0.05), bandMedium},
		{ptr(0.1), bandHigh},
		{ptr(0.49), bandHigh},
		{ptr(0.5), bandCritical},
		{ptr(0.97), bandCritical},
	}
	for _, c := range cases {
		if got := epssBand(c.in); got != c.want {
			t.Errorf("epssBand(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestBandLabel(t *testing.T) {
	cases := map[band]string{
		bandCritical: "CRITICAL",
		bandHigh:     "HIGH",
		bandMedium:   "MEDIUM",
		bandLow:      "LOW",
		bandNone:     naMarker,
	}
	for b, want := range cases {
		if got := bandLabel(b); got != want {
			t.Errorf("bandLabel(%v) = %q, want %q", b, got, want)
		}
	}
}

func TestCVSSString(t *testing.T) {
	if got := cvssString(nil); got != naMarker {
		t.Errorf("cvssString(nil) = %q, want %q", got, naMarker)
	}
	if got := cvssString(ptr(9.8)); got != "9.8" {
		t.Errorf("cvssString(9.8) = %q, want 9.8", got)
	}
	if got := cvssString(ptr(10)); got != "10.0" {
		t.Errorf("cvssString(10) = %q, want 10.0", got)
	}
}

func TestEPSSString(t *testing.T) {
	if got := epssString(nil); got != naMarker {
		t.Errorf("epssString(nil) = %q, want %q", got, naMarker)
	}
	if got := epssString(ptr(0.5)); got != "50.0%" {
		t.Errorf("epssString(0.5) = %q, want 50.0%%", got)
	}
	if got := epssString(ptr(0.001)); got != "0.1%" {
		t.Errorf("epssString(0.001) = %q, want 0.1%%", got)
	}
}

func TestRelativeAge(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	cases := []struct {
		unix int64
		want string
	}{
		{0, naMarker},
		{now.Unix() + 100, "just now"},
		{now.Unix() - 30, "just now"},
		{now.Unix() - 120, "2m ago"},
		{now.Unix() - 7200, "2h ago"},
		{now.Unix() - 172800, "2d ago"},
	}
	for _, c := range cases {
		if got := relativeAge(c.unix, now); got != c.want {
			t.Errorf("relativeAge(%d) = %q, want %q", c.unix, got, c.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	cases := []struct {
		in   string
		max  int
		want string
	}{
		{"hello world", 20, "hello world"},
		{"hello world", 5, "hell" + ellipsis},
		{"  multi   space ", 20, "multi space"},
		{"abc", 1, ellipsis},
		{"abc", 0, ""},
	}
	for _, c := range cases {
		if got := truncate(c.in, c.max); got != c.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", c.in, c.max, got, c.want)
		}
	}
}

func TestClusterSignals(t *testing.T) {
	c := store.DigestCluster{
		CVEs: []store.DigestCVE{
			{ID: "CVE-A", CVSSScore: ptr(7.5), EPSS: ptr(0.2)},
			{ID: "CVE-B", CVSSScore: ptr(9.8), EPSS: ptr(0.9), IsKEV: true},
			{ID: "CVE-C"},
		},
	}
	if !clusterHasKEV(c) {
		t.Error("clusterHasKEV = false, want true")
	}
	if got := clusterMaxCVSS(c); got == nil || *got != 9.8 {
		t.Errorf("clusterMaxCVSS = %v, want 9.8", got)
	}
	if got := clusterMaxEPSS(c); got == nil || *got != 0.9 {
		t.Errorf("clusterMaxEPSS = %v, want 0.9", got)
	}
	if got := clusterBand(c); got != bandCritical {
		t.Errorf("clusterBand = %v, want bandCritical (KEV)", got)
	}

	noKev := store.DigestCluster{CVEs: []store.DigestCVE{{ID: "CVE-D", CVSSScore: ptr(7.5)}}}
	if got := clusterBand(noKev); got != bandHigh {
		t.Errorf("clusterBand(no kev, 7.5) = %v, want bandHigh", got)
	}
	empty := store.DigestCluster{}
	if got := clusterBand(empty); got != bandNone {
		t.Errorf("clusterBand(empty) = %v, want bandNone", got)
	}
}

func TestOutletColor(t *testing.T) {
	cases := map[int]string{1: colorDim, 2: colorBlue, 3: colorCyan, 4: colorViolet, 9: colorViolet}
	for n, want := range cases {
		if got := outletColor(n); got != want {
			t.Errorf("outletColor(%d) = %q, want %q", n, got, want)
		}
	}
}
