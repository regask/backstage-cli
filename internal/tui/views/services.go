package views

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
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
	st := table.DefaultStyles()
	st.Selected = theme.Selected
	st.Header = theme.TableHeader
	t := table.New(table.WithColumns([]table.Column{
		{Title: "SERVICE", Width: 24},
		{Title: "ENV", Width: 12},
		{Title: "VERSION", Width: 20},
		{Title: "SYNC", Width: 10},
		{Title: "HEALTH", Width: 10},
	}), table.WithFocused(true), table.WithStyles(st))
	return Services{theme: theme, keys: keys, table: t, filter: ti}
}

// SetSize records the body height allotted to this view (App.View's total
// screen height minus its 1-line header and 1-line footer) and sizes the
// table to fill it, minus the lines this view's own View() renders on top.
func (s Services) SetSize(w, h int) Services {
	s.w, s.h = w, h
	return s.resize()
}

// resize applies s.w/s.h to the table, accounting for the "N services"
// header line View() always renders, plus the filter line when it's open —
// called both from SetSize and whenever filterOn toggles, since that changes
// how many lines View() reserves for itself.
func (s Services) resize() Services {
	s.table.SetWidth(s.w)
	reserved := 1 // "N services" header line
	if s.filterOn {
		reserved++ // filter input line
	}
	if bh := s.h - reserved; bh > 0 {
		s.table.SetHeight(bh)
	} else {
		s.table.SetHeight(0)
	}
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
			sync := s.theme.StatusStyle(d.SyncStatus).Render(dash(d.SyncStatus))
			health := s.theme.StatusStyle(d.HealthStatus).Render(dash(d.HealthStatus))
			trows = append(trows, table.Row{name, e, dash(d.Tag), sync, health})
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
		// A service with no envs emits zero table rows (see applyFilter); it
		// contributes nothing to the cursor mapping and is never selectable.
		rowsForSvc := len(r.Envs)
		if cur < n+rowsForSvc {
			return r, true
		}
		n += rowsForSvc
	}
	return s.shown[len(s.shown)-1], true
}

// FilterActive reports whether the "/" filter input is currently capturing
// keystrokes, so App can route input straight to it and skip the global and
// action key bindings (which would otherwise fire on filter text like "q" or
// "p" inside a service name).
func (s Services) FilterActive() bool { return s.filterOn }

func (s Services) Update(msg tea.Msg) (Services, tea.Cmd) {
	var cmd tea.Cmd
	if s.filterOn {
		switch m := msg.(type) {
		case tea.KeyMsg:
			switch m.String() {
			case "enter", "esc":
				s.filterOn = false
				s.filter.Blur()
				return s.resize().applyFilter(), nil
			}
		}
		s.filter, cmd = s.filter.Update(msg)
		return s.applyFilter(), cmd
	}
	if km, ok := msg.(tea.KeyMsg); ok && key.Matches(km, s.keys.Filter) {
		s.filterOn = true
		s.filter.Focus()
		return s.resize(), textinput.Blink
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
