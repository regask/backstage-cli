package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// commandBar is a `:`-prompted line for switching views or quitting, in the
// style of a vim command line.
type commandBar struct{ input textinput.Model }

func newCommandBar() commandBar {
	ti := textinput.New()
	ti.Prompt = ":"
	ti.Placeholder = "services | approvals | quit"
	return commandBar{input: ti}
}

func (c *commandBar) Focus() { c.input.Focus() }
func (c *commandBar) Blur() {
	c.input.SetValue("")
	c.input.Blur()
}
func (c commandBar) Focused() bool { return c.input.Focused() }

func (c commandBar) Update(msg tea.Msg) (commandBar, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.Type {
		case tea.KeyEnter:
			cmd := c.run(strings.TrimSpace(c.input.Value()))
			c.Blur()
			return c, cmd
		case tea.KeyEsc:
			c.Blur()
			return c, nil
		}
	}
	var cmd tea.Cmd
	c.input, cmd = c.input.Update(msg)
	return c, cmd
}

func (c commandBar) run(cmdText string) tea.Cmd {
	switch cmdText {
	case "services", "approvals":
		v := cmdText
		return func() tea.Msg { return switchViewMsg{View: v} }
	case "q", "quit":
		return tea.Quit
	default:
		t := fmt.Sprintf("unknown command: %q", cmdText)
		return func() tea.Msg { return bannerMsg{Text: t, IsErr: true} }
	}
}

func (c commandBar) View() string { return c.input.View() }
