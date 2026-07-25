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

	// Panel is a rounded-border card; callers set BorderForeground per-use
	// (e.g. semantic status color) via Panel.BorderForeground(...).
	Panel lipgloss.Style
	// Label/Value style a "key: value" line inside a detail pane — Label
	// muted, Value bold in the normal foreground.
	Label lipgloss.Style
	Value lipgloss.Style

	// Badge* render a solid-background status pill (see StatusBadge). Kept as
	// fields rather than computed inline so StatusBadge stays a one-line
	// switch.
	BadgeGood  lipgloss.Style
	BadgeWarn  lipgloss.Style
	BadgeBad   lipgloss.Style
	BadgeMuted lipgloss.Style

	// BrandColor/GoodColor/WarnColor/BadColor/MutedColor are the raw semantic
	// colors backing Title/Good/Warn/Bad/Muted, exposed for callers that need
	// a color rather than a full Style (e.g. Panel.BorderForeground).
	BrandColor lipgloss.AdaptiveColor
	GoodColor  lipgloss.AdaptiveColor
	WarnColor  lipgloss.AdaptiveColor
	BadColor   lipgloss.AdaptiveColor
	MutedColor lipgloss.AdaptiveColor

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

	// Badge pills use a solid, non-adaptive background — the pill supplies its
	// own contrast independent of the terminal's light/dark mode, so a single
	// saturated color (rather than a Light/Dark pair) is enough. Amber reads
	// better with dark text; green/red/muted get white.
	white := lipgloss.Color("#ffffff")
	onAmber := lipgloss.Color("#1a1a1a")
	badgeGoodBg := lipgloss.Color("#15803d")
	badgeWarnBg := lipgloss.Color("#d97706")
	badgeBadBg := lipgloss.Color("#b91c1c")
	badgeMutedBg := lipgloss.Color("#6b7280")

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

		Panel: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1),
		Label: lipgloss.NewStyle().Foreground(muted),
		Value: lipgloss.NewStyle().Bold(true).Foreground(fg),

		BadgeGood:  lipgloss.NewStyle().Bold(true).Padding(0, 1).Foreground(white).Background(badgeGoodBg),
		BadgeWarn:  lipgloss.NewStyle().Bold(true).Padding(0, 1).Foreground(onAmber).Background(badgeWarnBg),
		BadgeBad:   lipgloss.NewStyle().Bold(true).Padding(0, 1).Foreground(white).Background(badgeBadBg),
		BadgeMuted: lipgloss.NewStyle().Bold(true).Padding(0, 1).Foreground(white).Background(badgeMutedBg),

		BrandColor: brand,
		GoodColor:  good,
		WarnColor:  warn,
		BadColor:   bad,
		MutedColor: muted,

		BarBg: barBg,
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

// StatusColor is StatusStyle's color alone, for callers that need to paint
// something other than text with it (e.g. a Panel border).
func (t Theme) StatusColor(v string) lipgloss.AdaptiveColor {
	switch v {
	case "Synced", "Healthy":
		return t.GoodColor
	case "OutOfSync", "Progressing", "Missing", "Suspended":
		return t.WarnColor
	case "Degraded", "Unknown":
		return t.BadColor
	default:
		return t.MutedColor
	}
}

// StatusBadge renders v as a padded, solid-background pill, colored by the
// same sync/health semantics as StatusStyle/StatusColor.
func (t Theme) StatusBadge(v string) string {
	switch v {
	case "Synced", "Healthy":
		return t.BadgeGood.Render(v)
	case "OutOfSync", "Progressing", "Missing", "Suspended":
		return t.BadgeWarn.Render(v)
	case "Degraded", "Unknown":
		return t.BadgeBad.Render(v)
	default:
		return t.BadgeMuted.Render(v)
	}
}

// StatusGlyph is a compact, colorless marker for the same sync/health
// semantics, safe to embed inside a bubbles/table cell: table cells are
// plain text (an ANSI color code embedded in a cell value throws off
// runewidth.Truncate's column-width math, per table.Model.renderRow), so
// status there is conveyed by symbol rather than color. StatusBadge carries
// the full color version for detail panes, where free text is safe.
func (t Theme) StatusGlyph(v string) string {
	switch v {
	case "Synced", "Healthy":
		return "✓"
	case "OutOfSync", "Progressing", "Missing", "Suspended":
		return "~"
	case "Degraded", "Unknown":
		return "✗"
	default:
		return "-"
	}
}
