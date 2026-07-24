package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	bannerErr    bool // true when banner is an error (errMsg/bannerMsg{IsErr:true}), styled Bad instead of Good
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
		bodyH := m.Height - 2 // 1 line header + 1 line footer
		if bodyH < 0 {
			bodyH = 0
		}
		a.services = a.services.SetSize(m.Width, bodyH)
		a.approvals = a.approvals.SetSize(m.Width, bodyH)
		return a, nil
	case switchViewMsg:
		a.active = m.View
		a.banner = ""
		a.bannerErr = false
		return a, a.refresh()
	case servicesLoadedMsg:
		a.services = a.services.SetRows(m.Rows)
		return a, nil
	case approvalsLoadedMsg:
		a.approvals = a.approvals.SetItems(m.Items)
		return a, nil
	case bannerMsg:
		a.banner = m.Text
		a.bannerErr = m.IsErr
		return a, nil
	case errMsg:
		a.banner = m.Err.Error()
		a.bannerErr = true
		return a, nil
	case actionResultMsg:
		a.banner = m.Text
		a.bannerErr = !m.OK
		return a, a.refresh()
	case tea.KeyMsg:
		if a.cmdBar.Focused() {
			a.cmdBar, cmd = a.cmdBar.Update(msg)
			return a, cmd
		}
		// While the help overlay is up, only its own toggle and quit act;
		// every other key is swallowed so it can't move the underlying
		// table/selection.
		if a.showHelp {
			switch {
			case key.Matches(m, a.keys.Help):
				a.showHelp = false
				return a, nil
			case key.Matches(m, a.keys.Quit):
				return a, tea.Quit
			}
			return a, nil
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
				if strings.TrimSpace(env) == "" {
					// Keep the prompt open rather than launch a scaffolder
					// task with an empty environment.
					return a, nil
				}
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
		// While a view is capturing input (services filter open, approvals
		// detail viewport open), every key belongs to it: a service name
		// containing "q"/"p"/"r"/":" would otherwise quit/promote/release/
		// open the command bar, and approve/reject shouldn't fire against a
		// hidden approvals list underneath the detail pane. Skip the global
		// and action key switch entirely and fall through to delegate below;
		// only ctrl+c remains a hard-quit escape hatch.
		capturing := (a.active == "services" && (a.services.FilterActive() || a.services.DetailActive())) ||
			(a.active == "approvals" && a.approvals.DetailActive())
		if capturing {
			if m.Type == tea.KeyCtrlC {
				return a, tea.Quit
			}
		} else {
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
			case key.Matches(m, a.keys.SwitchView):
				if a.active == "approvals" {
					a.active = "services"
				} else {
					a.active = "approvals"
				}
				a.banner = ""
				a.bannerErr = false
				return a, a.refresh()
			case key.Matches(m, a.keys.Help):
				a.showHelp = !a.showHelp
				return a, nil
			case key.Matches(m, a.keys.Refresh):
				a.banner = ""
				a.bannerErr = false
				return a, a.refresh()
			case key.Matches(m, a.keys.Quit):
				return a, tea.Quit
			}
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
	// Too small to render meaningfully (and h-2 would go negative for the
	// header/footer split) — render nothing rather than a garbled frame.
	if a.h <= 2 || a.w <= 2 {
		return ""
	}
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
	footer := renderFooter(a.theme, a.keys, a.banner, a.bannerErr)
	switch {
	case a.cmdBar.Focused():
		footer = a.cmdBar.View()
	case a.confirmText != "":
		footer = a.theme.Banner.Render(a.confirmText)
	case a.promptActive:
		footer = a.prompt.View()
	}
	out := header + "\n" + body + "\n" + footer
	// Constrain the composed frame to exactly the terminal's dimensions so
	// an oversized view (from a view rendering more lines than its allotted
	// body height) never reaches the alt-screen renderer and causes redraw
	// thrash — pad short frames, truncate long ones, in both dimensions.
	return lipgloss.NewStyle().Width(a.w).MaxWidth(a.w).Height(a.h).MaxHeight(a.h).Render(out)
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
