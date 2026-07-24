package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/regask/backstage-cli/internal/contracts"
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

// TestAppApproveConfirmFlow: pressing "a" on approvals sets a confirm
// banner and stashes the pending command; any non-y key cancels it; "y"
// runs it and the resulting actionResultMsg sets the banner.
func TestAppApproveConfirmFlow(t *testing.T) {
	a := testApp()
	a.cl = newFakeClient(`{}`)
	m, _ := a.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	a = m.(App)
	m, _ = a.Update(switchViewMsg{View: "approvals"})
	a = m.(App)
	a.approvals = a.approvals.SetItems([]contracts.ApprovalRequest{{ID: "abc", Kind: "k", Status: "pending", Title: "t"}})

	m, _ = a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	a = m.(App)
	if a.confirmText == "" || a.pending == nil {
		t.Fatalf("expected pending confirm, got confirmText=%q pending=%v", a.confirmText, a.pending)
	}
	if !strings.Contains(a.View(), "approve abc") {
		t.Fatalf("confirm banner not shown in view: %s", a.View())
	}

	// A non-y key cancels.
	m, cmd := a.Update(tea.KeyMsg{Type: tea.KeyEsc})
	a = m.(App)
	if a.confirmText != "" || a.pending != nil || cmd != nil {
		t.Fatalf("expected cancel, got confirmText=%q pending=%v cmd=%v", a.confirmText, a.pending, cmd)
	}

	// Re-trigger and confirm with y.
	m, _ = a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	a = m.(App)
	m, cmd = a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	a = m.(App)
	if a.confirmText != "" || a.pending != nil {
		t.Fatalf("expected confirm cleared, got confirmText=%q pending=%v", a.confirmText, a.pending)
	}
	if cmd == nil {
		t.Fatalf("expected a command to run")
	}
	msg := cmd()
	res, ok := msg.(actionResultMsg)
	if !ok || !res.OK {
		t.Fatalf("want ok actionResultMsg, got %#v", msg)
	}

	m, _ = a.Update(res)
	a = m.(App)
	if a.banner != res.Text {
		t.Fatalf("banner = %q, want %q", a.banner, res.Text)
	}
}

// TestAppPromoteEnvPromptFlow: pressing "p" on services activates the env
// prompt; typing a value and pressing Enter runs promoteCmd with the typed
// env and the selected service ref.
func TestAppPromoteEnvPromptFlow(t *testing.T) {
	a := testApp()
	a.cl = newFakeClient(`{"id":"task-1"}`)
	m, _ := a.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	a = m.(App)
	a.services = a.services.SetRows([]contracts.MatrixRow{
		{ServiceRef: "component:default/alpha", ServiceName: "alpha",
			Envs: map[string]contracts.EnvDeploy{"production": {Tag: "v1"}}},
	})

	m, _ = a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	a = m.(App)
	if !a.promptActive || a.promptKind != "promote" || a.promptService != "component:default/alpha" {
		t.Fatalf("prompt not armed: active=%v kind=%q service=%q", a.promptActive, a.promptKind, a.promptService)
	}

	for _, r := range "staging" {
		m, _ = a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		a = m.(App)
	}
	m, cmd := a.Update(tea.KeyMsg{Type: tea.KeyEnter})
	a = m.(App)
	if a.promptActive {
		t.Fatalf("expected prompt cleared")
	}
	if cmd == nil {
		t.Fatalf("expected a command to run")
	}
	msg := cmd()
	res, ok := msg.(actionResultMsg)
	if !ok || !res.OK {
		t.Fatalf("want ok actionResultMsg, got %#v", msg)
	}
}
