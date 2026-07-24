package views

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/regask/backstage-cli/internal/contracts"
	"github.com/regask/backstage-cli/internal/tui/ui"
)

// Services is the deploy-matrix view: one table row per (service, env), with
// a "/"-toggled filter over service name/ref.
type Services struct {
	theme    ui.Theme
	keys     ui.Keys
	table    table.Model
	filter   textinput.Model
	filterOn bool
	rows     []contracts.MatrixRow // full, unfiltered
	shown    []contracts.MatrixRow // after filter (parallel to table rows)
	w, h     int
}

func NewServices(theme ui.Theme, keys ui.Keys) Services {
	ti := textinput.New()
	ti.Placeholder = "filter services…"
	ti.Prompt = "/"
	t := table.New(table.WithColumns([]table.Column{
		{Title: "SERVICE", Width: 24},
		{Title: "ENV", Width: 12},
		{Title: "VERSION", Width: 20},
		{Title: "SYNC", Width: 10},
		{Title: "HEALTH", Width: 10},
	}), table.WithFocused(true))
	return Services{theme: theme, keys: keys, table: t, filter: ti}
}

func (s Services) SetSize(w, h int) Services {
	s.w, s.h = w, h
	s.table.SetWidth(w)
	s.table.SetHeight(h - 3)
	return s
}

func (s Services) SetRows(rows []contracts.MatrixRow) Services {
	s.rows = rows
	return s.applyFilter()
}

func (s Services) applyFilter() Services {
	q := strings.ToLower(strings.TrimSpace(s.filter.Value()))
	s.shown = s.shown[:0]
	var trows []table.Row
	for _, r := range s.rows {
		if q != "" && !strings.Contains(strings.ToLower(r.ServiceName+" "+r.ServiceRef), q) {
			continue
		}
		s.shown = append(s.shown, r)
		envs := make([]string, 0, len(r.Envs))
		for e := range r.Envs {
			envs = append(envs, e)
		}
		sort.Strings(envs)
		// One table row per (service,env); first env carries the service name.
		for i, e := range envs {
			name := r.ServiceName
			if i > 0 {
				name = ""
			}
			d := r.Envs[e]
			trows = append(trows, table.Row{name, e, dash(d.Tag), dash(d.SyncStatus), dash(d.HealthStatus)})
		}
	}
	s.table.SetRows(trows)
	return s
}

func dash(v string) string {
	if v == "" {
		return "-"
	}
	return v
}

// Selected returns the service under the cursor (maps table cursor → shown[]).
func (s Services) Selected() (contracts.MatrixRow, bool) {
	if len(s.shown) == 0 {
		return contracts.MatrixRow{}, false
	}
	// Walk shown[] counting emitted rows until we reach the table cursor.
	cur := s.table.Cursor()
	n := 0
	for _, r := range s.shown {
		rowsForSvc := len(r.Envs)
		if rowsForSvc == 0 {
			rowsForSvc = 1
		}
		if cur < n+rowsForSvc {
			return r, true
		}
		n += rowsForSvc
	}
	return s.shown[len(s.shown)-1], true
}

func (s Services) Update(msg tea.Msg) (Services, tea.Cmd) {
	var cmd tea.Cmd
	if s.filterOn {
		switch m := msg.(type) {
		case tea.KeyMsg:
			switch m.String() {
			case "enter", "esc":
				s.filterOn = false
				s.filter.Blur()
				return s.applyFilter(), nil
			}
		}
		s.filter, cmd = s.filter.Update(msg)
		return s.applyFilter(), cmd
	}
	if km, ok := msg.(tea.KeyMsg); ok && km.String() == "/" {
		s.filterOn = true
		s.filter.Focus()
		return s, textinput.Blink
	}
	s.table, cmd = s.table.Update(msg)
	return s, cmd
}

func (s Services) View() string {
	head := s.theme.TableHeader.Render(fmt.Sprintf(" %d services ", len(s.shown)))
	body := s.table.View()
	if s.filterOn {
		return head + "\n" + s.filter.View() + "\n" + body
	}
	return head + "\n" + body
}
