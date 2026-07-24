package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/regask/backstage-cli/internal/contracts"
	"github.com/regask/backstage-cli/internal/tui/ui"
)

// Approvals is the approval-requests view: a list table with an Enter-toggled
// detail pane showing the release link and originating task backlink.
type Approvals struct {
	theme   ui.Theme
	keys    ui.Keys
	table   table.Model
	detail  viewport.Model
	showing bool
	items   []contracts.ApprovalRequest
	w, h    int
}

func NewApprovals(theme ui.Theme, keys ui.Keys) Approvals {
	t := table.New(table.WithColumns([]table.Column{
		{Title: "KIND", Width: 20},
		{Title: "STATUS", Width: 12},
		{Title: "TITLE", Width: 40},
	}), table.WithFocused(true))
	return Approvals{theme: theme, keys: keys, table: t, detail: viewport.New(0, 0)}
}

func (a Approvals) SetSize(w, h int) Approvals {
	a.w, a.h = w, h
	a.table.SetWidth(w)
	if bh := h - 3; bh > 0 {
		a.table.SetHeight(bh)
	}
	a.detail.Width = w
	a.detail.Height = h - 3
	return a
}

func (a Approvals) SetItems(items []contracts.ApprovalRequest) Approvals {
	a.items = items
	rows := make([]table.Row, 0, len(items))
	for _, it := range items {
		rows = append(rows, table.Row{it.Kind, it.Status, it.Title})
	}
	a.table.SetRows(rows)
	return a
}

// Selected returns the approval under the cursor (1:1 with the table rows,
// unlike Services which fans one row out per env).
func (a Approvals) Selected() (contracts.ApprovalRequest, bool) {
	if len(a.items) == 0 {
		return contracts.ApprovalRequest{}, false
	}
	i := a.table.Cursor()
	if i < 0 || i >= len(a.items) {
		i = 0
	}
	return a.items[i], true
}

// detailText mirrors cmd/approvals_detail.go's printApprovalDetail: the
// release link prefers resultUrl (published) and falls back to the draft
// release link(s) derived from the payload while the approval is still
// pending, plus the originating scaffolder task backlink.
func (a Approvals) detailText(r contracts.ApprovalRequest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s  [%s]  status=%s\n", r.Title, r.Kind, r.Status)
	if r.Requester != "" {
		fmt.Fprintf(&b, "requested by %s\n", r.Requester)
	}
	if r.Summary != "" {
		fmt.Fprintf(&b, "\n%s\n", r.Summary)
	}
	if r.ResultURL != "" {
		fmt.Fprintf(&b, "\nrelease link: %s\n", r.ResultURL)
	} else if drafts := r.DraftReleaseURLs(); len(drafts) > 0 {
		if len(drafts) == 1 {
			fmt.Fprintf(&b, "\ndraft release: %s\n", drafts[0])
		} else {
			fmt.Fprintf(&b, "\ndraft releases:\n")
			for _, d := range drafts {
				fmt.Fprintf(&b, "  %s\n", d)
			}
		}
	}
	if task := r.TaskID(); task != "" {
		fmt.Fprintf(&b, "task: %s\n", task)
	}
	if r.Error != "" {
		fmt.Fprintf(&b, "\nerror: %s\n", r.Error)
	}
	return b.String()
}

func (a Approvals) Update(msg tea.Msg) (Approvals, tea.Cmd) {
	var cmd tea.Cmd
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "enter":
			if sel, ok := a.Selected(); ok {
				a.detail.SetContent(a.detailText(sel))
				a.showing = true
				return a, nil
			}
		case "esc":
			if a.showing {
				a.showing = false
				return a, nil
			}
		}
	}
	if a.showing {
		a.detail, cmd = a.detail.Update(msg)
		return a, cmd
	}
	a.table, cmd = a.table.Update(msg)
	return a, cmd
}

func (a Approvals) View() string {
	if a.showing {
		return a.detail.View()
	}
	head := a.theme.TableHeader.Render(fmt.Sprintf(" %d approvals ", len(a.items)))
	return head + "\n" + a.table.View()
}
