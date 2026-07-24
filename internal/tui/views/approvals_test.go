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

	if !strings.Contains(view, "draft release: https://github.com/regask/other-svc/releases/tag/v2") {
		t.Fatalf("view missing draft release link: %s", view)
	}
	if !strings.Contains(view, "task: task-123") {
		t.Fatalf("view missing task backlink: %s", view)
	}
}
