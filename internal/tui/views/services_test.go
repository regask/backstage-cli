package views

import (
	"testing"

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
	s.SetSize(80, 24)
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
