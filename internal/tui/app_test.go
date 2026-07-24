package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"strings"
	"testing"
)

func testApp() App { return NewApp(nil, "https://p", "user:default/me") }

func TestAppSwitchViewAndBanner(t *testing.T) {
	a := testApp()
	m, _ := a.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	a = m.(App)

	m, _ = a.Update(switchViewMsg{View: "approvals"})
	a = m.(App)
	if a.active != "approvals" {
		t.Fatalf("active = %q", a.active)
	}

	m, _ = a.Update(errMsg{Err: errString("boom")})
	a = m.(App)
	if !strings.Contains(a.View(), "boom") {
		t.Fatalf("banner not shown in view")
	}
}

type errString string

func (e errString) Error() string { return string(e) }
