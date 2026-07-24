package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/regask/backstage-cli/internal/contracts"
	"github.com/regask/backstage-cli/internal/tui/ui"
)

func newRows() []contracts.MatrixRow {
	return []contracts.MatrixRow{
		{ServiceRef: "component:default/alpha", ServiceName: "alpha",
			Envs: map[string]contracts.EnvDeploy{
				"development": {Tag: "v1-dev", SyncStatus: "Synced", HealthStatus: "Healthy"},
				"production":  {Tag: "v1", SyncStatus: "Synced", HealthStatus: "Healthy"},
			}},
		{ServiceRef: "component:default/beta", ServiceName: "beta",
			Envs: map[string]contracts.EnvDeploy{"production": {Tag: "v2"}}},
	}
}

// A service now emits exactly one table row, with one column per env, instead
// of the old one-row-per-(service,env) fan-out.
func TestServicesMatrixOneRowPerService(t *testing.T) {
	s := NewServices(ui.NewTheme(), ui.DefaultKeys())
	s = s.SetSize(120, 24)
	s = s.SetRows(newRows())

	if got := len(s.table.Rows()); got != 2 {
		t.Fatalf("table rows = %d, want 2", got)
	}

	sel, ok := s.Selected()
	if !ok || sel.ServiceName != "alpha" {
		t.Fatalf("selected = %+v ok=%v, want alpha", sel, ok)
	}

	out := s.View()
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") {
		t.Fatalf("view missing service names: %q", out)
	}
	if !strings.Contains(out, "v1-dev") {
		t.Fatalf("view missing alpha's DEV tag: %q", out)
	}
	if !strings.Contains(out, "v1") {
		t.Fatalf("view missing alpha's PROD tag: %q", out)
	}
	if !strings.Contains(out, "v2") {
		t.Fatalf("view missing beta's PROD tag: %q", out)
	}
}

// Selected maps the table cursor straight to shown[] — no more fan-out
// row-counting, since every service is exactly one row.
func TestServicesSelectedSimple(t *testing.T) {
	s := NewServices(ui.NewTheme(), ui.DefaultKeys())
	s = s.SetSize(120, 24)
	s = s.SetRows(newRows())

	s.table.SetCursor(1)
	sel, ok := s.Selected()
	if !ok || sel.ServiceName != "beta" {
		t.Fatalf("selected = %+v ok=%v, want beta", sel, ok)
	}
}

// FilterActive must flip on once "/" is pressed, so App can route keystrokes
// straight to the filter input instead of the global/action key bindings.
func TestServicesFilterActive(t *testing.T) {
	s := NewServices(ui.NewTheme(), ui.DefaultKeys())
	s = s.SetSize(80, 24)
	if s.FilterActive() {
		t.Fatal("filter should not be active before '/' is pressed")
	}

	s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	if !s.FilterActive() {
		t.Fatal("filter should be active after '/' is pressed")
	}
}

// Enter opens the per-service detail pane, which is where sync/health color
// lives (the table itself stays plain text for column alignment).
func TestServicesDetailShowsPerEnv(t *testing.T) {
	s := NewServices(ui.NewTheme(), ui.DefaultKeys())
	s = s.SetSize(120, 24)
	s = s.SetRows(newRows())

	s, _ = s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !s.DetailActive() {
		t.Fatal("detail should be active after enter")
	}

	out := s.View()
	if !strings.Contains(out, "development") {
		t.Fatalf("detail missing env name: %q", out)
	}
	if !strings.Contains(out, "v1-dev") {
		t.Fatalf("detail missing env tag: %q", out)
	}
	if !strings.Contains(out, "sync=") || !strings.Contains(out, "health=") {
		t.Fatalf("detail missing sync/health: %q", out)
	}

	s, _ = s.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if s.DetailActive() {
		t.Fatal("detail should close on esc")
	}
}

func TestDash(t *testing.T) {
	if dash("") != "-" {
		t.Fatal(`dash("") should be "-"`)
	}
	if dash("v1") != "v1" {
		t.Fatal(`dash("v1") should be "v1"`)
	}
}
