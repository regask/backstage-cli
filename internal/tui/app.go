package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/regask/backstage-cli/internal/client"
	"github.com/regask/backstage-cli/internal/tui/ui"
	"github.com/regask/backstage-cli/internal/tui/views"
)

// App is the root Bubble Tea model: it owns the command bar, the header/
// footer chrome, and the two views (services/approvals), dispatching by
// active view and handling the global key bindings.
type App struct {
	cl           *client.Client
	portal, user string
	theme        ui.Theme
	keys         ui.Keys
	cmdBar       commandBar
	services     views.Services
	approvals    views.Approvals
	active       string
	banner       string
	showHelp     bool
	w, h         int

	// confirm gates a destructive action (approve/reject) behind a y/N
	// prompt; pending is the command it runs once confirmed.
	confirmText string
	pending     tea.Cmd

	// prompt captures the target environment for promote/release.
	prompt        textinput.Model
	promptActive  bool
	promptKind    string // "promote" | "release"
	promptService string
}

func NewApp(cl *client.Client, portal, user string) App {
	theme, keys := ui.NewTheme(), ui.DefaultKeys()
	prompt := textinput.New()
	prompt.Prompt = "target env: "
	return App{
		cl: cl, portal: portal, user: user, theme: theme, keys: keys,
		cmdBar:    newCommandBar(),
		services:  views.NewServices(theme, keys),
		approvals: views.NewApprovals(theme, keys),
		active:    "services",
		prompt:    prompt,
	}
}

func (a App) Init() tea.Cmd { return loadServices(a.cl, false) }

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		a.w, a.h = m.Width, m.Height
		bodyH := m.Height - 2
		a.services = a.services.SetSize(m.Width, bodyH)
		a.approvals = a.approvals.SetSize(m.Width, bodyH)
		return a, nil
	case switchViewMsg:
		a.active = m.View
		a.banner = ""
		return a, a.refresh()
	case servicesLoadedMsg:
		a.services = a.services.SetRows(m.Rows)
		return a, nil
	case approvalsLoadedMsg:
		a.approvals = a.approvals.SetItems(m.Items)
		return a, nil
	case bannerMsg:
		a.banner = m.Text
		return a, nil
	case errMsg:
		a.banner = m.Err.Error()
		return a, nil
	case actionResultMsg:
		a.banner = m.Text
		return a, a.refresh()
	case tea.KeyMsg:
		if a.cmdBar.Focused() {
			a.cmdBar, cmd = a.cmdBar.Update(msg)
			return a, cmd
		}
		// Confirm takes precedence over everything except the command bar:
		// any key other than y/Y cancels the pending action.
		if a.confirmText != "" {
			if m.String() == "y" || m.String() == "Y" {
				cmd = a.pending
				a.pending = nil
				a.confirmText = ""
				return a, cmd
			}
			a.pending = nil
			a.confirmText = ""
			return a, nil
		}
		if a.promptActive {
			switch m.Type {
			case tea.KeyEnter:
				env := a.prompt.Value()
				a.promptActive = false
				a.prompt.Blur()
				a.prompt.SetValue("")
				switch a.promptKind {
				case "promote":
					return a, promoteCmd(a.cl, env, []string{a.promptService})
				case "release":
					return a, releaseCmd(a.cl, env)
				}
				return a, nil
			case tea.KeyEsc:
				a.promptActive = false
				a.prompt.Blur()
				a.prompt.SetValue("")
				return a, nil
			}
			a.prompt, cmd = a.prompt.Update(msg)
			return a, cmd
		}
		switch {
		case a.active == "approvals" && key.Matches(m, a.keys.Approve):
			if sel, ok := a.approvals.Selected(); ok {
				a.confirmText = "approve " + sel.ID + "? [y/N]"
				a.pending = approveCmd(a.cl, sel.ID, true)
			}
			return a, nil
		case a.active == "approvals" && key.Matches(m, a.keys.Reject):
			if sel, ok := a.approvals.Selected(); ok {
				a.confirmText = "reject " + sel.ID + "? [y/N]"
				a.pending = approveCmd(a.cl, sel.ID, false)
			}
			return a, nil
		case a.active == "services" && key.Matches(m, a.keys.Promote):
			if sel, ok := a.services.Selected(); ok {
				a.promptKind = "promote"
				a.promptService = sel.ServiceRef
				a.promptActive = true
				a.prompt.Focus()
				return a, textinput.Blink
			}
			return a, nil
		case a.active == "services" && key.Matches(m, a.keys.Release):
			a.promptService = ""
			if sel, ok := a.services.Selected(); ok {
				a.promptService = sel.ServiceRef
			}
			a.promptKind = "release"
			a.promptActive = true
			a.prompt.Focus()
			return a, textinput.Blink
		case key.Matches(m, a.keys.Command):
			a.cmdBar.Focus()
			return a, nil
		case key.Matches(m, a.keys.Help):
			a.showHelp = !a.showHelp
			return a, nil
		case key.Matches(m, a.keys.Refresh):
			a.banner = ""
			return a, a.refresh()
		case key.Matches(m, a.keys.Quit):
			return a, tea.Quit
		}
	}
	// delegate to the active view
	switch a.active {
	case "approvals":
		a.approvals, cmd = a.approvals.Update(msg)
	default:
		a.services, cmd = a.services.Update(msg)
	}
	return a, cmd
}

func (a App) refresh() tea.Cmd {
	if a.active == "approvals" {
		return loadApprovals(a.cl, true)
	}
	return loadServices(a.cl, true)
}

func (a App) View() string {
	header := renderHeader(a.theme, a.portal, a.user, a.active)
	var body string
	switch a.active {
	case "approvals":
		body = a.approvals.View()
	default:
		body = a.services.View()
	}
	if a.showHelp {
		body = a.theme.Modal.Render(helpText(a.keys))
	}
	footer := renderFooter(a.theme, a.keys, a.banner)
	switch {
	case a.cmdBar.Focused():
		footer = a.cmdBar.View()
	case a.confirmText != "":
		footer = a.theme.Banner.Render(a.confirmText)
	case a.promptActive:
		footer = a.prompt.View()
	}
	return header + "\n" + body + "\n" + footer
}

func helpText(k ui.Keys) string {
	var b string
	for _, grp := range k.FullHelp() {
		for _, bind := range grp {
			b += bind.Help().Key + "  " + bind.Help().Desc + "\n"
		}
		b += "\n"
	}
	return b
}
