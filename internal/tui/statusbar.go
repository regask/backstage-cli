package tui

import (
	"fmt"
	"strings"

	"github.com/regask/backstage-cli/internal/tui/ui"
)

func renderHeader(theme ui.Theme, portal, user, view string) string {
	tabs := []string{"services", "approvals"}
	var parts []string
	for _, t := range tabs {
		if t == view {
			parts = append(parts, theme.TabActive.Render(t))
		} else {
			parts = append(parts, theme.TabInactive.Render(t))
		}
	}
	left := theme.Title.Render("bsr") + "  " + strings.Join(parts, "  ")
	right := theme.Muted.Render(fmt.Sprintf("%s  %s", user, portal))
	return theme.Header.Render(left + "   " + right)
}

func renderFooter(theme ui.Theme, keys ui.Keys, banner string, bannerErr bool) string {
	if banner != "" {
		style := theme.Good
		if bannerErr {
			style = theme.Bad
		}
		return style.Bold(true).Padding(0, 1).Render(banner)
	}
	var hints []string
	for _, b := range keys.ShortHelp() {
		hints = append(hints, b.Help().Key+" "+b.Help().Desc)
	}
	return theme.Footer.Render(strings.Join(hints, " · "))
}
