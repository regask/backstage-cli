package ui

import "github.com/charmbracelet/lipgloss"

// Theme is the single source of visual styling, adaptive to light/dark.
type Theme struct {
	Title       lipgloss.Style
	TabActive   lipgloss.Style
	TabInactive lipgloss.Style
	Header      lipgloss.Style
	Footer      lipgloss.Style
	TableHeader lipgloss.Style
	Selected    lipgloss.Style
	Banner      lipgloss.Style
	Modal       lipgloss.Style
	Good        lipgloss.Style
	Warn        lipgloss.Style
	Bad         lipgloss.Style
	Muted       lipgloss.Style

	// BarBg is the header/footer chrome-bar background. It's exposed as a raw
	// color (not just baked into Header/Footer) because the bars are
	// hand-composed from several independently-rendered pieces (title, tabs,
	// key hints, literal spacers) — lipgloss can't carry a wrapper's
	// Background through the ANSI reset that terminates each nested Render
	// call, so every piece must carry this same background itself.
	BarBg lipgloss.AdaptiveColor
}

func NewTheme() Theme {
	fg := lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#e6e6e6"}
	muted := lipgloss.AdaptiveColor{Light: "#6b7280", Dark: "#9ca3af"}
	brand := lipgloss.AdaptiveColor{Light: "#2563eb", Dark: "#60a5fa"}
	good := lipgloss.AdaptiveColor{Light: "#15803d", Dark: "#4ade80"}
	warn := lipgloss.AdaptiveColor{Light: "#b45309", Dark: "#fbbf24"}
	bad := lipgloss.AdaptiveColor{Light: "#b91c1c", Dark: "#f87171"}
	barBg := lipgloss.AdaptiveColor{Light: "#e5e7eb", Dark: "#1c1f26"}
	selBg := lipgloss.AdaptiveColor{Light: "#dbeafe", Dark: "#1e3a5f"}
	selFg := lipgloss.AdaptiveColor{Light: "#1e3a8a", Dark: "#dbeafe"}

	return Theme{
		Title:       lipgloss.NewStyle().Bold(true).Foreground(brand).Background(barBg),
		TabActive:   lipgloss.NewStyle().Bold(true).Foreground(brand).Underline(true).Background(barBg),
		TabInactive: lipgloss.NewStyle().Foreground(muted).Background(barBg),
		Header:      lipgloss.NewStyle().Foreground(fg).Background(barBg).Padding(0, 1),
		Footer:      lipgloss.NewStyle().Foreground(muted).Background(barBg).Padding(0, 1),
		TableHeader: lipgloss.NewStyle().Bold(true).Foreground(muted),
		Selected:    lipgloss.NewStyle().Bold(true).Foreground(selFg).Background(selBg),
		Banner:      lipgloss.NewStyle().Foreground(bad).Bold(true).Padding(0, 1),
		Modal:       lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(brand).Padding(1, 2),
		Good:        lipgloss.NewStyle().Foreground(good),
		Warn:        lipgloss.NewStyle().Foreground(warn),
		Bad:         lipgloss.NewStyle().Foreground(bad),
		Muted:       lipgloss.NewStyle().Foreground(muted),
		BarBg:       barBg,
	}
}

// StatusStyle picks a semantic style for an argocd sync/health value.
func (t Theme) StatusStyle(v string) lipgloss.Style {
	switch v {
	case "Synced", "Healthy":
		return t.Good
	case "OutOfSync", "Progressing", "Missing", "Suspended":
		return t.Warn
	case "Degraded", "Unknown":
		return t.Bad
	default:
		return t.Muted
	}
}
