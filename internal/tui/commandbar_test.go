package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCommandBarSwitchView(t *testing.T) {
	cb := newCommandBar()
	cb.Focus()
	// type "approvals"
	for _, r := range "approvals" {
		cb, _ = cb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	_, cmd := cb.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should emit a command")
	}
	msg := cmd()
	if sv, ok := msg.(switchViewMsg); !ok || sv.View != "approvals" {
		t.Fatalf("want switchViewMsg{approvals}, got %T %+v", msg, msg)
	}
}
