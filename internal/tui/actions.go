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

// TODO(execution): confirm template:default/{promote-code,release} against
// the running backend, same as the non-TUI CLI commands.
func promoteCmd(cl *client.Client, toEnv string, services []string) tea.Cmd {
	return func() tea.Msg {
		values := map[string]any{"toEnvironment": toEnv, "services": services}
		id, err := scaffolder.Launch(context.Background(), cl, "template:default/promote-code", values)
		if err != nil {
			return errMsg{err}
		}
		return actionResultMsg{OK: true, Text: "promote launched: task " + id}
	}
}

func releaseCmd(cl *client.Client, env string) tea.Cmd {
	return func() tea.Msg {
		values := map[string]any{"environment": env}
		id, err := scaffolder.Launch(context.Background(), cl, "template:default/release", values)
		if err != nil {
			return errMsg{err}
		}
		return actionResultMsg{OK: true, Text: "release launched: task " + id}
	}
}
