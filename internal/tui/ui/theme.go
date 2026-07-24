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
}

func NewTheme() Theme {
	fg := lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#e6e6e6"}
	muted := lipgloss.AdaptiveColor{Light: "#6b7280", Dark: "#9ca3af"}
	brand := lipgloss.AdaptiveColor{Light: "#2563eb", Dark: "#60a5fa"}
	good := lipgloss.AdaptiveColor{Light: "#15803d", Dark: "#4ade80"}
	warn := lipgloss.AdaptiveColor{Light: "#b45309", Dark: "#fbbf24"}
	bad := lipgloss.AdaptiveColor{Light: "#b91c1c", Dark: "#f87171"}

	return Theme{
		Title:       lipgloss.NewStyle().Bold(true).Foreground(brand),
		TabActive:   lipgloss.NewStyle().Bold(true).Foreground(brand).Underline(true),
		TabInactive: lipgloss.NewStyle().Foreground(muted),
		Header:      lipgloss.NewStyle().Foreground(fg).Padding(0, 1),
		Footer:      lipgloss.NewStyle().Foreground(muted).Padding(0, 1),
		TableHeader: lipgloss.NewStyle().Bold(true).Foreground(muted),
		Selected:    lipgloss.NewStyle().Bold(true).Foreground(brand),
		Banner:      lipgloss.NewStyle().Foreground(bad).Bold(true).Padding(0, 1),
		Modal:       lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(brand).Padding(1, 2),
		Good:        lipgloss.NewStyle().Foreground(good),
		Warn:        lipgloss.NewStyle().Foreground(warn),
		Bad:         lipgloss.NewStyle().Foreground(bad),
		Muted:       lipgloss.NewStyle().Foreground(muted),
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
