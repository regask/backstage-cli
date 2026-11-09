package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/regask/backstage-cli/internal/client"
	"github.com/regask/backstage-cli/internal/scaffolder"
)

type actionResultMsg struct {
	OK   bool
	Text string
}

func approveCmd(cl *client.Client, id string, approve bool) tea.Cmd {
	return func() tea.Msg {
		if err := cl.DecideApproval(context.Background(), id, approve); err != nil {
			return errMsg{err}
		}
		verb := "approved"
		if !approve {
			verb = "rejected"
		}
		return actionResultMsg{OK: true, Text: fmt.Sprintf("request %s %s", id, verb)}
	}
}

func promoteCmd(cl *client.Client, toEnv string, services []string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		values, err := scaffolder.PromoteValues(ctx, cl, toEnv, services, "")
		if err != nil {
			return errMsg{err}
		}
		id, err := scaffolder.Launch(ctx, cl, scaffolder.PromoteTemplate, values)
		if err != nil {
			return errMsg{err}
		}
		return actionResultMsg{OK: true, Text: "promote launched: task " + id}
	}
}

func releaseCmd(cl *client.Client, env string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		values, err := scaffolder.ReleaseValues(ctx, cl, env, nil, 0)
		if err != nil {
			return errMsg{err}
		}
		id, err := scaffolder.Launch(ctx, cl, scaffolder.ReleaseTemplate, values)
		if err != nil {
			return errMsg{err}
		}
		return actionResultMsg{OK: true, Text: "release launched: task " + id}
	}
}
