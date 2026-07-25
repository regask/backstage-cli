package ui

import "github.com/charmbracelet/bubbles/key"

// Keys are the global key bindings shared across the app and views.
type Keys struct {
	Up         key.Binding
	Down       key.Binding
	Enter      key.Binding
	Filter     key.Binding
	Command    key.Binding
	SwitchView key.Binding
	Refresh    key.Binding
	Help       key.Binding
	Back       key.Binding
	Quit       key.Binding
	// View-scoped action keys (handled by the active view).
	Promote key.Binding
	Release key.Binding
	Approve key.Binding
	Reject  key.Binding
}

func DefaultKeys() Keys {
	return Keys{
		Up:         key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:       key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Enter:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open")),
		Filter:     key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
		Command:    key.NewBinding(key.WithKeys(":"), key.WithHelp(":", "command")),
		SwitchView: key.NewBinding(key.WithKeys("tab", "shift+tab"), key.WithHelp("tab", "switch view")),
		Refresh:    key.NewBinding(key.WithKeys("ctrl+r"), key.WithHelp("ctrl+r", "refresh")),
		Help:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Back:       key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		Quit:       key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Promote:    key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "promote")),
		Release:    key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "release")),
		Approve:    key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "approve")),
		Reject:     key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "reject")),
	}
}

func (k Keys) ShortHelp() []key.Binding {
	return []key.Binding{k.Command, k.SwitchView, k.Filter, k.Refresh, k.Help, k.Quit}
}

func (k Keys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.Back},
		{k.Filter, k.Command, k.SwitchView, k.Refresh},
		{k.Promote, k.Release, k.Approve, k.Reject},
		{k.Help, k.Quit},
	}
}
