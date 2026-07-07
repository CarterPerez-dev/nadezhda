// ©AngelaMos | 2026
// theme.go

package tui

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	colorVoid  = "#0B0620"
	colorText  = "#E7E0FF"
	colorMuted = "#9A8AD0"
	colorDim   = "#574B7D"

	colorMagenta = "#FF2E97"
	colorAmber   = "#FF8A3D"
	colorYellow  = "#FFE04D"
	colorCyan    = "#00E5FF"
	colorViolet  = "#A65CFF"
	colorBlue    = "#2D6BFF"
)

const (
	barFull  = "█"
	barEmpty = "░"
	ruleLead = "▓▒░"
	ruleBody = "─"
)

func bandColor(b band) string {
	switch b {
	case bandCritical:
		return colorMagenta
	case bandHigh:
		return colorAmber
	case bandMedium:
		return colorYellow
	case bandLow:
		return colorCyan
	default:
		return colorDim
	}
}

type Theme struct {
	Brand       lipgloss.Style
	BrandMark   lipgloss.Style
	Meta        lipgloss.Style
	ColHead     lipgloss.Style
	Text        lipgloss.Style
	Muted       lipgloss.Style
	Dim         lipgloss.Style
	Rank        lipgloss.Style
	RankSel     lipgloss.Style
	Headline    lipgloss.Style
	HeadlineSel lipgloss.Style
	SelBar      lipgloss.Style
	KeyGlyph    lipgloss.Style
	KeyDesc     lipgloss.Style
	Spinner     lipgloss.Style
	PanelTitle  lipgloss.Style
	Label       lipgloss.Style
	Link        lipgloss.Style
}

func NewTheme() Theme {
	base := lipgloss.NewStyle().Foreground(lipgloss.Color(colorText))
	return Theme{
		Brand:       lipgloss.NewStyle().Foreground(lipgloss.Color(colorViolet)).Bold(true),
		BrandMark:   lipgloss.NewStyle().Foreground(lipgloss.Color(colorMagenta)).Bold(true),
		Meta:        lipgloss.NewStyle().Foreground(lipgloss.Color(colorCyan)),
		ColHead:     lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted)).Bold(true),
		Text:        base,
		Muted:       lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted)),
		Dim:         lipgloss.NewStyle().Foreground(lipgloss.Color(colorDim)),
		Rank:        lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted)),
		RankSel:     lipgloss.NewStyle().Foreground(lipgloss.Color(colorMagenta)).Bold(true),
		Headline:    base,
		HeadlineSel: lipgloss.NewStyle().Foreground(lipgloss.Color(colorText)).Bold(true),
		SelBar:      lipgloss.NewStyle().Foreground(lipgloss.Color(colorMagenta)).Bold(true),
		KeyGlyph:    lipgloss.NewStyle().Foreground(lipgloss.Color(colorCyan)).Bold(true),
		KeyDesc:     lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted)),
		Spinner:     lipgloss.NewStyle().Foreground(lipgloss.Color(colorMagenta)).Bold(true),
		PanelTitle:  lipgloss.NewStyle().Foreground(lipgloss.Color(colorViolet)).Bold(true),
		Label:       lipgloss.NewStyle().Foreground(lipgloss.Color(colorBlue)).Bold(true),
		Link:        lipgloss.NewStyle().Foreground(lipgloss.Color(colorCyan)).Underline(true),
	}
}

func (t Theme) fg(color string) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color))
}

func (t Theme) bandFG(b band) lipgloss.Style {
	return t.fg(bandColor(b))
}

func (t Theme) chip(text, color string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorVoid)).
		Background(lipgloss.Color(color)).
		Bold(true).
		Render(" " + text + " ")
}

func (t Theme) spectrumBar(frac float64, width int) string {
	if width <= 0 {
		return ""
	}
	filled := int(math.Round(clamp01(frac) * float64(width)))
	stops := []string{colorCyan, colorBlue, colorViolet, colorMagenta}
	var b strings.Builder
	for i := 0; i < width; i++ {
		if i < filled {
			b.WriteString(t.fg(stops[i*len(stops)/width]).Render(barFull))
		} else {
			b.WriteString(t.Dim.Render(barEmpty))
		}
	}
	return b.String()
}

func (t Theme) bandBar(b band, frac float64, width int) string {
	if width <= 0 {
		return ""
	}
	filled := int(math.Round(clamp01(frac) * float64(width)))
	color := bandColor(b)
	var sb strings.Builder
	for i := 0; i < width; i++ {
		if i < filled {
			sb.WriteString(t.fg(color).Render(barFull))
		} else {
			sb.WriteString(t.Dim.Render(barEmpty))
		}
	}
	return sb.String()
}

func (t Theme) rule(width int) string {
	lead := len([]rune(ruleLead)) + 1
	n := width - lead
	if n < 0 {
		return t.fg(colorViolet).Render(ruleLead)
	}
	return t.fg(colorViolet).Render(ruleLead) + " " + t.Dim.Render(strings.Repeat(ruleBody, n))
}
