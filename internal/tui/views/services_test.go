package views

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/regask/backstage-cli/internal/contracts"
	"github.com/regask/backstage-cli/internal/tui/ui"
)

func newRows() []contracts.MatrixRow {
	return []contracts.MatrixRow{
		{ServiceRef: "component:default/alpha", ServiceName: "alpha",
			Envs: map[string]contracts.EnvDeploy{"production": {Tag: "v1", SyncStatus: "Synced", HealthStatus: "Healthy"}}},
		{ServiceRef: "component:default/beta", ServiceName: "beta",
			Envs: map[string]contracts.EnvDeploy{"production": {Tag: "v2"}}},
	}
}

func TestServicesSetRowsAndSelect(t *testing.T) {
	s := NewServices(ui.NewTheme(), ui.DefaultKeys())
	s = s.SetSize(80, 24)
	s = s.SetRows(newRows())
	sel, ok := s.Selected()
	if !ok || sel.ServiceName != "alpha" {
		t.Fatalf("selected = %+v ok=%v", sel, ok)
	}
	out := s.View()
	if out == "" {
		t.Fatal("view should render")
	}
}

// A service with an empty Envs map (undeployed) emits zero table rows, so it
// must never be counted when mapping the cursor back to a shown[] entry.
func TestServicesSelectedSkipsEmptyEnvs(t *testing.T) {
	rows := []contracts.MatrixRow{
		{ServiceRef: "component:default/undeployed", ServiceName: "undeployed",
			Envs: map[string]contracts.EnvDeploy{}},
		{ServiceRef: "component:default/beta", ServiceName: "beta",
			Envs: map[string]contracts.EnvDeploy{"production": {Tag: "v2"}}},
	}
	s := NewServices(ui.NewTheme(), ui.DefaultKeys())
	s = s.SetSize(80, 24)
	s = s.SetRows(rows)

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
