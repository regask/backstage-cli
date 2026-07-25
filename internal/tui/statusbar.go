package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/regask/backstage-cli/internal/tui/ui"
)

// barText renders literal chrome-bar text (spacers, separators) with the
// bar's background. Every visible run inside a hand-composed bar needs its
// own copy of the background — lipgloss can't propagate a wrapper's
// Background through the ANSI reset that ends each nested Render() call, so
// a plain, unstyled string sitting between two colored segments would show
// through with no background at all.
func barText(theme ui.Theme, s string) string {
	return lipgloss.NewStyle().Background(theme.BarBg).Render(s)
}

// fillBar composes left and right into a single line exactly w columns wide:
// left, a background-colored spacer, then right pinned to the far edge.
// left and right must already carry the bar's background on every visible
// run (see barText) and must already fit within w (see fitBar) — this only
// fills the space between them.
func fillBar(theme ui.Theme, w int, left, right string) string {
	gap := w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}
	spacer := lipgloss.NewStyle().Background(theme.BarBg).Width(gap).Render("")
	return lipgloss.JoinHorizontal(lipgloss.Top, left, spacer, right)
}

// fitBar truncates right, then left as a last resort, so left+right never
// exceeds w — otherwise an overlong bar (a long user/portal string, a narrow
// terminal) would overflow past w and get word-wrapped onto a second line by
// the outer frame clamp in App.View, breaking the one-line-exactly-w-wide
// guarantee. ansi.Truncate is escape-code aware, so cutting mid-string can't
// corrupt the embedded color codes.
func fitBar(w int, left, right string) (string, string) {
	if budget := w - lipgloss.Width(left); lipgloss.Width(right) > budget {
		if budget < 0 {
			budget = 0
		}
		right = ansi.Truncate(right, budget, "…")
	}
	if lipgloss.Width(left) > w {
		left = ansi.Truncate(left, w, "…")
	}
	return left, right
}

// renderHeader renders the top chrome as a full-width, single-line bar: "bsr"
// plus the view tabs on the left, the user/portal on the right. w is the
// terminal width; theme.Header's Padding(0,1) contributes the 1-column inset
// on each side, so the hand-composed content targets w-2. approvalsCount, when
// positive, renders as a small attention pill after the "approvals" tab label
// so the pending count is visible from any view, not just while on that tab.
func renderHeader(theme ui.Theme, w int, portal, user, view string, approvalsCount int) string {
	tabs := []string{"services", "approvals"}
	var parts []string
	for _, t := range tabs {
		var rendered string
		if t == view {
			rendered = theme.TabActive.Render(t)
		} else {
			rendered = theme.TabInactive.Render(t)
		}
		if t == "approvals" && approvalsCount > 0 {
			rendered += theme.CountBadge.Render(fmt.Sprintf(" %d ", approvalsCount))
		}
		parts = append(parts, rendered)
	}
	left := theme.Title.Render("bsr") + barText(theme, "  ") + strings.Join(parts, barText(theme, "  "))
	right := theme.Muted.Background(theme.BarBg).Render(fmt.Sprintf("%s  %s", user, portal))

	inner := w - 2
	if inner < 0 {
		inner = 0
	}
	left, right = fitBar(inner, left, right)
	return theme.Header.Render(fillBar(theme, inner, left, right))
}

// renderFooter renders the bottom chrome as a full-width, single-line bar:
// either the key hints (bold key + muted description, · separated) or, when
// a banner is set, the banner text styled semantically (red for an error,
// green for success) in its place.
func renderFooter(theme ui.Theme, w int, keys ui.Keys, banner string, bannerErr bool) string {
	inner := w - 2
	if inner < 0 {
		inner = 0
	}

	var content string
	if banner != "" {
		style := theme.Good
		if bannerErr {
			style = theme.Bad
		}
		content = style.Bold(true).Background(theme.BarBg).Render(banner)
	} else {
		var hints []string
		for _, b := range keys.ShortHelp() {
			h := b.Help()
			key := theme.Title.Render(h.Key)
			desc := theme.Muted.Background(theme.BarBg).Render(h.Desc)
			hints = append(hints, key+barText(theme, " ")+desc)
		}
		content = strings.Join(hints, barText(theme, " · "))
	}
	content, _ = fitBar(inner, content, "")
	return theme.Footer.Render(fillBar(theme, inner, content, ""))
}
