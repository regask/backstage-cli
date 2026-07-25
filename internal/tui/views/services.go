package views

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/regask/backstage-cli/internal/contracts"
	"github.com/regask/backstage-cli/internal/tui/ui"
)

// matrixEnvs is the fixed column order for the deploy matrix, mapped to the
// backend's env keys.
var matrixEnvs = []struct{ label, key string }{
	{"DEV", "development"},
	{"STG", "staging"},
	{"PRE-PROD", "pre-prod"},
	{"PROD", "production"},
}

// Services is the deploy-matrix view: one table row per service, one column
// per environment, with a "/"-toggled filter over service name/ref and an
// Enter-toggled detail pane for the selected service's per-env sync/health.
type Services struct {
	theme    ui.Theme
	keys     ui.Keys
	table    table.Model
	filter   textinput.Model
	filterOn bool
	detail   viewport.Model
	showing  bool
	rows     []contracts.MatrixRow // full, unfiltered
	shown    []contracts.MatrixRow // after filter (1:1 with table rows)
	w, h     int
}

func NewServices(theme ui.Theme, keys ui.Keys) Services {
	ti := textinput.New()
	ti.Placeholder = "filter services…"
	ti.Prompt = "/"
	// Start from table.DefaultStyles() and only override color/bold — never
	// touch padding. Header and Cell must keep matching horizontal padding or
	// the header text drifts from its column data (cumulative, worst on the
	// rightmost column). Selected wraps the whole already-padded row, so it's
	// safe to replace outright.
	st := table.DefaultStyles()
	st.Header = st.Header.Bold(true).Foreground(theme.TableHeader.GetForeground())
	st.Selected = theme.Selected
	cols := []table.Column{{Title: "SERVICE", Width: 28}}
	for _, e := range matrixEnvs {
		cols = append(cols, table.Column{Title: e.label, Width: 16})
	}
	t := table.New(table.WithColumns(cols), table.WithFocused(true), table.WithStyles(st))
	return Services{theme: theme, keys: keys, table: t, filter: ti, detail: viewport.New(0, 0)}
}

// SetSize records the body height allotted to this view (App.View's total
// screen height minus its 1-line header and 1-line footer) and sizes the
// table/detail viewport to fill it, minus the lines this view's own View()
// renders on top.
func (s Services) SetSize(w, h int) Services {
	s.w, s.h = w, h
	s.detail.Width = w
	if h > 0 {
		s.detail.Height = h
	} else {
		s.detail.Height = 0
	}
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
		row := table.Row{r.ServiceName}
		for _, e := range matrixEnvs {
			row = append(row, dash(r.Envs[e.key].Tag))
		}
		trows = append(trows, row)
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

// Selected returns the service under the cursor (1:1 with the table rows —
// unlike the old fan-out layout, one table row is always exactly one
// service).
func (s Services) Selected() (contracts.MatrixRow, bool) {
	if len(s.shown) == 0 {
		return contracts.MatrixRow{}, false
	}
	i := s.table.Cursor()
	if i < 0 || i >= len(s.shown) {
		i = 0
	}
	return s.shown[i], true
}

// FilterActive reports whether the "/" filter input is currently capturing
// keystrokes, so App can route input straight to it and skip the global and
// action key bindings (which would otherwise fire on filter text like "q" or
// "p" inside a service name).
func (s Services) FilterActive() bool { return s.filterOn }

// DetailActive reports whether the Enter detail viewport is open, so App can
// route input (scrolling, esc) straight to it and skip global/action key
// bindings that would otherwise fire against the hidden table underneath.
func (s Services) DetailActive() bool { return s.showing }

// detailText renders the selected service's ref plus, for each env it's
// deployed to (sorted), its tag and sync/health colored via
// theme.StatusStyle — free text in a viewport, so ANSI color is fine here
// (unlike the table, where embedding it breaks column alignment).
func (s Services) detailText(r contracts.MatrixRow) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", r.ServiceRef)
	envs := make([]string, 0, len(r.Envs))
	for e := range r.Envs {
		envs = append(envs, e)
	}
	sort.Strings(envs)
	for _, e := range envs {
		d := r.Envs[e]
		sync := s.theme.StatusStyle(d.SyncStatus).Render(dash(d.SyncStatus))
		health := s.theme.StatusStyle(d.HealthStatus).Render(dash(d.HealthStatus))
		fmt.Fprintf(&b, "%s: %s  sync=%s  health=%s\n", e, dash(d.Tag), sync, health)
		if d.ArgocdURL != "" {
			fmt.Fprintf(&b, "  %s\n", d.ArgocdURL)
		}
	}
	return b.String()
}

func (s Services) Update(msg tea.Msg) (Services, tea.Cmd) {
	var cmd tea.Cmd
	if s.showing {
		if km, ok := msg.(tea.KeyMsg); ok && km.String() == "esc" {
			s.showing = false
			return s, nil
		}
		s.detail, cmd = s.detail.Update(msg)
		return s, cmd
	}
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
	if km, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(km, s.keys.Filter) {
			s.filterOn = true
			s.filter.Focus()
			return s.resize(), textinput.Blink
		}
		if km.String() == "enter" {
			if sel, ok := s.Selected(); ok {
				s.detail.SetContent(s.detailText(sel))
				s.showing = true
				return s, nil
			}
		}
	}
	s.table, cmd = s.table.Update(msg)
	return s, cmd
}

func (s Services) View() string {
	if s.showing {
		return s.detail.View()
	}
	head := s.theme.TableHeader.Render(fmt.Sprintf(" %d services ", len(s.shown)))
	body := s.table.View()
	if s.filterOn {
		return head + "\n" + s.filter.View() + "\n" + body
	}
	return head + "\n" + body
}
