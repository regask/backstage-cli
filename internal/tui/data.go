package tui

import (
	"context"
	"net/url"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/regask/backstage-cli/internal/client"
	"github.com/regask/backstage-cli/internal/contracts"
)

// loadServices fetches the fleet matrix as a tea.Cmd.
func loadServices(cl *client.Client, fresh bool) tea.Cmd {
	return func() tea.Msg {
		rows, err := cl.Matrix(context.Background(), "", fresh)
		if err != nil {
			return errMsg{err}
		}
		return servicesLoadedMsg{Rows: rows}
	}
}

// loadApprovals fetches pending approvals. The list route returns
// { requests: [...] }.
func loadApprovals(cl *client.Client, fresh bool) tea.Cmd {
	return func() tea.Msg {
		var out struct {
			Requests []contracts.ApprovalRequest `json:"requests"`
		}
		q := url.Values{"status": {"pending"}}
		if err := cl.GetJSON(context.Background(), "/approvals/requests", q, fresh, &out); err != nil {
			return errMsg{err}
		}
		return approvalsLoadedMsg{Items: out.Requests}
	}
}
