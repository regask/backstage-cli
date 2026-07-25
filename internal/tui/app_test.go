package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
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

// TestAppRejectConfirmFlow: pressing "x" (reject) on approvals sets a confirm
// banner and stashes the pending command; "y" runs it and the resulting
// actionResultMsg sets the banner.
func TestAppRejectConfirmFlow(t *testing.T) {
	a := testApp()
	a.cl = newFakeClient(`{}`)
	m, _ := a.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	a = m.(App)
	m, _ = a.Update(switchViewMsg{View: "approvals"})
	a = m.(App)
	a.approvals = a.approvals.SetItems([]contracts.ApprovalRequest{{ID: "abc", Kind: "k", Status: "pending", Title: "t"}})

	m, _ = a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	a = m.(App)
	if a.confirmText == "" || a.pending == nil {
		t.Fatalf("expected pending confirm, got confirmText=%q pending=%v", a.confirmText, a.pending)
	}
	if !strings.Contains(a.View(), "reject abc") {
		t.Fatalf("confirm banner not shown in view: %s", a.View())
	}

	m, cmd := a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
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

// TestAppQuitKeyDuringServicesFilterDoesNotQuit: while the services "/"
// filter is capturing keystrokes, a "q" keystroke (typed as part of a service
// name) must reach the filter input, not the global quit binding. Feed the
// keys directly: the app must stay alive (no tea.QuitMsg command), the filter
// must stay active, and the "q" must land in the filter's rendered value.
func TestAppQuitKeyDuringServicesFilterDoesNotQuit(t *testing.T) {
	a := testApp()
	m, _ := a.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	a = m.(App)
	a.services = a.services.SetRows([]contracts.MatrixRow{
		{ServiceRef: "component:default/quickstart", ServiceName: "quickstart",
			Envs: map[string]contracts.EnvDeploy{"production": {Tag: "v1"}}},
	})

	m, _ = a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	a = m.(App)
	if !a.services.FilterActive() {
		t.Fatalf("expected filter to be active after '/'")
	}

	m, cmd := a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	a, ok := m.(App)
	if !ok {
		t.Fatalf("expected App to be returned, got %T", m)
	}
	if cmd != nil {
		if _, isQuit := cmd().(tea.QuitMsg); isQuit {
			t.Fatalf("expected 'q' during filter to not quit")
		}
	}
	if !a.services.FilterActive() {
		t.Fatalf("expected filter to remain active")
	}
	if !strings.Contains(a.View(), "/q") {
		t.Fatalf("expected 'q' to land in the filter input, view: %s", a.View())
	}
}

// TestAppTabSwitchesView: Tab is a global key that toggles the active view
// between services and approvals, without going through the command bar.
func TestAppTabSwitchesView(t *testing.T) {
	a := testApp()
	m, _ := a.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	a = m.(App)
	if a.active != "services" {
		t.Fatalf("active = %q, want services", a.active)
	}

	m, _ = a.Update(tea.KeyMsg{Type: tea.KeyTab})
	a = m.(App)
	if a.active != "approvals" {
		t.Fatalf("active = %q, want approvals", a.active)
	}

	m, _ = a.Update(tea.KeyMsg{Type: tea.KeyTab})
	a = m.(App)
	if a.active != "services" {
		t.Fatalf("active = %q, want services", a.active)
	}
}

// TestAppTabIgnoredDuringFilter: while the services "/" filter is capturing
// keystrokes, Tab must not be hijacked by the global switch-view binding —
// it belongs to the filter input (or is a no-op there), and the active view
// must stay unchanged.
func TestAppTabIgnoredDuringFilter(t *testing.T) {
	a := testApp()
	m, _ := a.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	a = m.(App)
	a.services = a.services.SetRows([]contracts.MatrixRow{
		{ServiceRef: "component:default/quickstart", ServiceName: "quickstart",
			Envs: map[string]contracts.EnvDeploy{"production": {Tag: "v1"}}},
	})

	m, _ = a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	a = m.(App)
	if !a.services.FilterActive() {
		t.Fatalf("expected filter to be active after '/'")
	}

	m, _ = a.Update(tea.KeyMsg{Type: tea.KeyTab})
	a = m.(App)
	if a.active != "services" {
		t.Fatalf("expected Tab during filter to not switch view, active = %q", a.active)
	}
}

// TestAppPromoteEmptyEnvKeepsPromptOpen: pressing Enter on the promote
// prompt with an empty (or whitespace-only) value must not launch anything;
// the prompt stays active for the user to type a value.
func TestAppPromoteEmptyEnvKeepsPromptOpen(t *testing.T) {
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
	if !a.promptActive {
		t.Fatalf("prompt not armed")
	}

	m, cmd := a.Update(tea.KeyMsg{Type: tea.KeyEnter})
	a = m.(App)
	if !a.promptActive {
		t.Fatalf("expected prompt to stay open on empty input")
	}
	if cmd != nil {
		t.Fatalf("expected no launch cmd for empty env, got %v", cmd)
	}
}

// hasBindingKey reports whether bindings contains one whose Help().Key
// matches want.
func hasBindingKey(bindings []key.Binding, want string) bool {
	for _, b := range bindings {
		if b.Help().Key == want {
			return true
		}
	}
	return false
}

// TestFooterHintsPerView: the footer must suggest the actions relevant to
// the active view — approve/reject on approvals, promote/release/filter on
// services — not a single static global list.
func TestFooterHintsPerView(t *testing.T) {
	a := testApp()

	a.active = "approvals"
	hints := a.footerHints()
	if !hasBindingKey(hints, a.keys.Approve.Help().Key) {
		t.Fatalf("approvals footer missing approve hint: %v", hints)
	}
	if !hasBindingKey(hints, a.keys.Reject.Help().Key) {
		t.Fatalf("approvals footer missing reject hint: %v", hints)
	}
	if hasBindingKey(hints, a.keys.Promote.Help().Key) {
		t.Fatalf("approvals footer should not advertise promote: %v", hints)
	}

	a.active = "services"
	hints = a.footerHints()
	if !hasBindingKey(hints, a.keys.Promote.Help().Key) {
		t.Fatalf("services footer missing promote hint: %v", hints)
	}
	if !hasBindingKey(hints, a.keys.Filter.Help().Key) {
		t.Fatalf("services footer missing filter hint: %v", hints)
	}
	if hasBindingKey(hints, a.keys.Approve.Help().Key) {
		t.Fatalf("services footer should not advertise approve: %v", hints)
	}
}
