package tui

import (
	"github.com/charmbracelet/bubbles/key"
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
}

func NewApp(cl *client.Client, portal, user string) App {
	theme, keys := ui.NewTheme(), ui.DefaultKeys()
	return App{
		cl: cl, portal: portal, user: user, theme: theme, keys: keys,
		cmdBar:    newCommandBar(),
		services:  views.NewServices(theme, keys),
		approvals: views.NewApprovals(theme, keys),
		active:    "services",
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
	case tea.KeyMsg:
		if a.cmdBar.Focused() {
			a.cmdBar, cmd = a.cmdBar.Update(msg)
			return a, cmd
		}
		switch {
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
	if a.cmdBar.Focused() {
		footer = a.cmdBar.View()
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
