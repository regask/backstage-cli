package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/regask/backstage-cli/internal/contracts"
	"github.com/regask/backstage-cli/internal/tui/ui"
)

func TestApprovalsSelectAndDetail(t *testing.T) {
	a := NewApprovals(ui.NewTheme(), ui.DefaultKeys())
	a = a.SetSize(80, 24)
	a = a.SetItems([]contracts.ApprovalRequest{
		{ID: "abc", Kind: "release-publish", Status: "pending", Title: "Publish svc",
			ResultURL: "https://gh/rel/v1", Summary: "publish"},
	})
	sel, ok := a.Selected()
	if !ok || sel.ID != "abc" {
		t.Fatalf("selected = %+v", sel)
	}
	if !strings.Contains(a.View(), "Publish svc") {
		t.Fatalf("view missing title: %s", a.View())
	}
}

// Pending release-publish approvals have no resultUrl yet; the detail view
// must fall back to the draft release link derived from the payload, and
// still show the task backlink.
func TestApprovalsDetailShowsDraftReleaseLink(t *testing.T) {
	a := NewApprovals(ui.NewTheme(), ui.DefaultKeys())
	a = a.SetSize(80, 24)
	a = a.SetItems([]contracts.ApprovalRequest{
		{ID: "def", Kind: "release-publish", Status: "pending", Title: "Publish other-svc",
			Requester: "louis", Summary: "publish other-svc v2",
			Payload: []byte(`{"owner":"regask","repo":"other-svc","tag":"v2","taskId":"task-123"}`)},
	})

	updated, _ := a.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view := updated.View()

	// Label and URL are styled separately (see detailText), so assert the
	// stable tokens rather than the old exact "label: value" concatenation.
	if !strings.Contains(view, "draft release") {
		t.Fatalf("view missing draft release label: %s", view)
	}
	if !strings.Contains(view, "https://github.com/regask/other-svc/releases/tag/v2") {
		t.Fatalf("view missing draft release link: %s", view)
	}
	if !strings.Contains(view, "task") || !strings.Contains(view, "task-123") {
		t.Fatalf("view missing task backlink: %s", view)
	}
}

// The detail pane renders status/kind as badges and the title in brand color
// (see ui.Theme.StatusBadge) rather than the old flat "title [kind] status="
// text — assert the rendered tokens survive.
func TestApprovalsDetailRendersBadgeAndTitle(t *testing.T) {
	a := NewApprovals(ui.NewTheme(), ui.DefaultKeys())
	a = a.SetSize(80, 24)
	a = a.SetItems([]contracts.ApprovalRequest{
		{ID: "ghi", Kind: "release-publish", Status: "pending", Title: "Publish third-svc",
			Requester: "louis", ResultURL: "https://gh/rel/v3"},
	})

	updated, _ := a.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view := updated.View()

	if !strings.Contains(view, "Publish third-svc") {
		t.Fatalf("view missing title: %s", view)
	}
	if !strings.Contains(view, "pending") {
		t.Fatalf("view missing status badge: %s", view)
	}
	if !strings.Contains(view, "release-publish") {
		t.Fatalf("view missing kind: %s", view)
	}
}
