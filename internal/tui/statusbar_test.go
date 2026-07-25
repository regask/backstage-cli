package tui

import (
	"strings"
	"testing"

	"github.com/regask/backstage-cli/internal/tui/ui"
)

// The header must surface the pending-approvals count as a badge next to the
// "approvals" tab, visible regardless of which view is active, so a user on
// "services" still sees there's something waiting on "approvals".
func TestHeaderShowsApprovalsCount(t *testing.T) {
	theme := ui.NewTheme()
	out := renderHeader(theme, 100, "portal", "user", "services", 3)
	if !strings.Contains(out, "3") {
		t.Fatalf("header missing approvals count badge: %q", out)
	}
}

// A zero count renders no badge — an empty/zero pill would just be noise
// next to the tab.
func TestHeaderHidesZeroApprovalsCount(t *testing.T) {
	theme := ui.NewTheme()
	out := renderHeader(theme, 100, "portal", "user", "services", 0)
	if strings.Contains(out, " 0 ") {
		t.Fatalf("header should not render a badge for a zero count: %q", out)
	}
}
