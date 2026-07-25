package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/regask/backstage-cli/internal/contracts"
	"github.com/regask/backstage-cli/internal/tui/ui"
)

// tableBorderRows/tableBorderCols are declared once in services.go and
// shared across this package's two list views.

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
	// Start from table.DefaultStyles() and only override color/bold — never
	// touch padding. Header and Cell must keep matching horizontal padding or
	// the header text drifts from its column data (cumulative, worst on the
	// rightmost column). Selected wraps the whole already-padded row in one
	// shot, so no per-cell style may carry its own color/border (see
	// services.go's NewServices for why) — status color is applied as a
	// post-render pass instead (see colorizeApprovalStatuses).
	st := table.DefaultStyles()
	st.Header = st.Header.Bold(true).Foreground(theme.TableHeader.GetForeground())
	st.Selected = theme.Selected
	t := table.New(table.WithColumns([]table.Column{
		{Title: "KIND", Width: 20},
		{Title: "STATUS", Width: 12},
		{Title: "TITLE", Width: 40},
	}), table.WithFocused(true), table.WithStyles(st))
	return Approvals{theme: theme, keys: keys, table: t, detail: viewport.New(0, 0)}
}

// SetSize records the body height allotted to this view (App.View's total
// screen height minus its 1-line header and 1-line footer). The table gets
// that height minus the "N approvals" header line and the bordered box
// View() wraps the table in; the detail viewport, which renders with no
// extra chrome, gets the full height.
func (a Approvals) SetSize(w, h int) Approvals {
	a.w, a.h = w, h
	a.table.SetWidth(w - tableBorderCols)
	reserved := 1 + tableBorderRows // "N approvals" header line + table border
	if bh := h - reserved; bh > 0 {
		a.table.SetHeight(bh)
	} else {
		a.table.SetHeight(0)
	}
	a.detail.Width = w
	if h > 0 {
		a.detail.Height = h
	} else {
		a.detail.Height = 0
	}
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

// approvalStatusStyle maps an approval's status word to a semantic style —
// a different vocabulary than services' argocd sync/health (see
// ui.Theme.StatusColor), so it's kept local to this view.
func approvalStatusStyle(theme ui.Theme, status string) lipgloss.Style {
	switch strings.ToLower(status) {
	case "approved", "published", "completed":
		return theme.Good
	case "pending", "in-progress", "in_progress":
		return theme.Warn
	case "rejected", "error", "failed", "expired":
		return theme.Bad
	default:
		return theme.Muted
	}
}

// colorizeApprovalStatuses recolors each row's STATUS word, bold + semantic
// color, as a post-render pass over the already-rendered table view — same
// reasoning as services.go's colorizeGlyphs: it runs after bubbles/table's
// own width truncation/padding, so it can't corrupt column alignment, and it
// locates each item's line by its (fairly unique) Title text rather than
// assuming a row index, since table.View() may only show a scrolled window
// of items. skipID is the currently-selected approval's ID, if any: its line
// is left untouched so the Selected wrap's background isn't cut short by a
// nested reset.
func colorizeApprovalStatuses(theme ui.Theme, view string, items []contracts.ApprovalRequest, skipID string) string {
	lines := strings.Split(view, "\n")
	for _, it := range items {
		if it.Title == "" || it.ID == skipID {
			continue
		}
		style := approvalStatusStyle(theme, it.Status)
		for i, line := range lines {
			if strings.Contains(line, it.Title) {
				lines[i] = strings.ReplaceAll(line, it.Status, style.Bold(true).Render(it.Status))
				break
			}
		}
	}
	return strings.Join(lines, "\n")
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
// pending, plus the originating scaffolder task backlink. Free text in a
// viewport, so ANSI color/borders are fine here (unlike the table).
func (a Approvals) detailText(r contracts.ApprovalRequest) string {
	title := a.theme.Value.Foreground(a.theme.BrandColor).Render(r.Title)
	sub := a.theme.StatusBadge(r.Status) + "  " + a.theme.BadgeMuted.Render(r.Kind)
	if r.Requester != "" {
		sub += "  " + a.theme.Muted.Render("requested by "+r.Requester)
	}

	width := a.detail.Width
	if width <= 0 {
		width = 80
	}

	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(sub)
	b.WriteString("\n")

	if r.Summary != "" {
		panelWidth := width - 2
		if panelWidth < 12 {
			panelWidth = 12
		}
		summary := a.theme.Panel.Width(panelWidth).BorderForeground(a.theme.MutedColor).Render(r.Summary)
		fmt.Fprintf(&b, "\n%s\n", summary)
	}

	linkWidth := width - len("release ")
	if r.ResultURL != "" {
		fmt.Fprintf(&b, "\n%s%s\n", a.theme.Label.Render("release "), ansi.Truncate(r.ResultURL, linkWidth, "…"))
	} else if drafts := r.DraftReleaseURLs(); len(drafts) > 0 {
		if len(drafts) == 1 {
			fmt.Fprintf(&b, "\n%s%s\n", a.theme.Label.Render("draft release "), ansi.Truncate(drafts[0], linkWidth, "…"))
		} else {
			fmt.Fprintf(&b, "\n%s\n", a.theme.Label.Render("draft releases"))
			for _, d := range drafts {
				fmt.Fprintf(&b, "  %s\n", ansi.Truncate(d, width-2, "…"))
			}
		}
	}
	if task := r.TaskID(); task != "" {
		fmt.Fprintf(&b, "\n%s%s\n", a.theme.Label.Render("task "), task)
	}
	if r.Error != "" {
		fmt.Fprintf(&b, "\n%s\n", a.theme.Bad.Render("error: "+r.Error))
	}
	return b.String()
}

// DetailActive reports whether the detail viewport is open, so App can route
// input (scrolling, esc) straight to it and skip global/action key bindings
// that would otherwise fire against the hidden list underneath.
func (a Approvals) DetailActive() bool { return a.showing }

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
	skipID := ""
	if sel, ok := a.Selected(); ok {
		skipID = sel.ID
	}
	tv := colorizeApprovalStatuses(a.theme, a.table.View(), a.items, skipID)
	body := a.theme.Panel.Padding(0, 0).BorderForeground(a.theme.MutedColor).Render(tv)
	return head + "\n" + body
}
