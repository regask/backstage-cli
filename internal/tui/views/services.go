package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
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
	// rightmost column). Selected wraps the whole already-padded row in one
	// shot (bubbles/table's renderRow joins every cell into a single plain
	// string first, then wraps *that* in Selected) — so Cell/Header carry no
	// color/border of their own: a per-cell style renders as its own
	// self-contained ANSI segment (with its own reset), which would cut the
	// outer Selected background off right after the first cell instead of
	// covering the whole row. Status color is instead applied as a
	// post-render pass over the finished table string (see colorizeGlyphs),
	// which runs after bubbles/table's own truncation/padding, so it can't
	// throw off column alignment or the row highlight.
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

// tableBorderRows/tableBorderCols are the outer rounded-border panel View()
// wraps the table in — reserved out of the allotted size before it reaches
// the table itself, same idea as the "N services" header line below.
const (
	tableBorderRows = 2 // top + bottom border
	tableBorderCols = 2 // left + right border
)

// resize applies s.w/s.h to the table, accounting for the "N services"
// header line View() always renders, the filter line when it's open, and the
// bordered box View() wraps the table in — called both from SetSize and
// whenever filterOn toggles, since that changes how many lines View()
// reserves for itself.
func (s Services) resize() Services {
	s.table.SetWidth(s.w - tableBorderCols)
	reserved := 1 // "N services" header line
	if s.filterOn {
		reserved++ // filter input line
	}
	reserved += tableBorderRows
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
			row = append(row, s.envCell(r.Envs[e.key]))
		}
		trows = append(trows, row)
	}
	s.table.SetRows(trows)
	return s
}

// envCell is the deploy-matrix table cell for one service/env: the tag plus
// a compact sync/health glyph pair. bubbles/table renders each cell as a
// single truncated, colorless line (see NewServices), so the two statuses
// ride along on the tag's line via ui.Theme.StatusGlyph rather than the
// colored badge/second line the detail pane can afford.
func (s Services) envCell(d contracts.EnvDeploy) string {
	if d.Tag == "" {
		return "-"
	}
	return fmt.Sprintf("%s %s%s", d.Tag, s.theme.StatusGlyph(d.SyncStatus), s.theme.StatusGlyph(d.HealthStatus))
}

func dash(v string) string {
	if v == "" {
		return "-"
	}
	return v
}

// colorizeGlyphs recolors the plain sync/health glyphs (✓/~/✗, see
// StatusGlyph) in an already-rendered table view, bold + semantic color. It
// runs as a decoration pass over the FINAL string, after bubbles/table has
// finished all its own width truncation/padding on the plain glyph — so
// unlike coloring the cell value up front, this can't corrupt column
// alignment. skip is the plain text of the currently-selected row (if any):
// that line is left untouched, since injecting a nested ANSI reset inside it
// would cut short the Selected wrap's own background/foreground.
func colorizeGlyphs(theme ui.Theme, view, skip string) string {
	if !strings.ContainsAny(view, "✓~✗") {
		return view
	}
	lines := strings.Split(view, "\n")
	for i, line := range lines {
		if skip != "" && strings.Contains(line, skip) {
			continue
		}
		line = strings.ReplaceAll(line, "✓", theme.Good.Bold(true).Render("✓"))
		line = strings.ReplaceAll(line, "~", theme.Warn.Bold(true).Render("~"))
		line = strings.ReplaceAll(line, "✗", theme.Bad.Bold(true).Render("✗"))
		lines[i] = line
	}
	return strings.Join(lines, "\n")
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

// detailText renders the selected service as a title (short name + full ref)
// followed by one rounded-border card per env it's deployed to (in
// matrixEnvs order), bordered in the env's health color — free text in a
// viewport, so ANSI color/borders are fine here (unlike the table, where
// embedding color breaks renderRow's column-width math).
func (s Services) detailText(r contracts.MatrixRow) string {
	short := strings.TrimPrefix(r.ServiceRef, "component:default/")
	title := s.theme.Value.Foreground(s.theme.BrandColor).Render(short)
	ref := s.theme.Muted.Render(r.ServiceRef)

	// Panel.Width sets the content box (padding included, border excluded —
	// see lipgloss's box model), so panelWidth+2 (border) lands on the
	// viewport's own width.
	panelWidth := s.detail.Width - 2
	if panelWidth < 12 {
		panelWidth = 12
	}
	urlWidth := panelWidth - 2 // minus Panel's own Padding(0,1)

	var cards []string
	for _, e := range matrixEnvs {
		d, ok := r.Envs[e.key]
		if !ok {
			continue
		}
		line1 := s.theme.Value.Render(strings.ToUpper(e.key)) + "  " + s.theme.Value.Render(dash(d.Tag))
		line2 := s.theme.Label.Render("sync ") + s.theme.StatusBadge(dash(d.SyncStatus)) +
			s.theme.Label.Render("   health ") + s.theme.StatusBadge(dash(d.HealthStatus))
		lines := []string{line1, line2}
		if d.ArgocdURL != "" {
			lines = append(lines, s.theme.Muted.Render(ansi.Truncate("↗ "+d.ArgocdURL, urlWidth, "…")))
		}
		card := s.theme.Panel.Width(panelWidth).
			BorderForeground(s.theme.StatusColor(d.HealthStatus)).
			Render(strings.Join(lines, "\n"))
		cards = append(cards, card)
	}

	return title + "\n" + ref + "\n\n" + strings.Join(cards, "\n") + "\n"
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
	skip := ""
	if sel, ok := s.Selected(); ok {
		skip = sel.ServiceName
	}
	tv := colorizeGlyphs(s.theme, s.table.View(), skip)
	body := s.theme.Panel.Padding(0, 0).BorderForeground(s.theme.MutedColor).Render(tv)
	if s.filterOn {
		return head + "\n" + s.filter.View() + "\n" + body
	}
	return head + "\n" + body
}
