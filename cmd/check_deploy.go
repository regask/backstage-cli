package cmd

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/regask/backstage-cli/internal/contracts"
	"github.com/regask/backstage-cli/internal/render"
	"github.com/spf13/cobra"
)

var checkDeployEnv string

// dash renders an empty cell as "-" for readable tables.
func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// matchesService accepts a bare name (alert-service), a full entity ref
// (component:default/alert-service), or a namespace/name — the backend matrix
// only exact-matches the full ref, so we match client-side to be forgiving.
func matchesService(row contracts.MatrixRow, q string) bool {
	ql := strings.ToLower(q)
	ref := strings.ToLower(row.ServiceRef)
	return strings.EqualFold(row.ServiceRef, q) ||
		strings.EqualFold(row.ServiceName, q) ||
		strings.HasSuffix(ref, "/"+ql)
}

func filterMatrixByName(rows []contracts.MatrixRow, q string) []contracts.MatrixRow {
	out := make([]contracts.MatrixRow, 0, len(rows))
	for _, row := range rows {
		if matchesService(row, q) {
			out = append(out, row)
		}
	}
	return out
}

// filterMatrixRows applies the --env filter to a matrix result, keeping it in
// sync between the table and --json renderings.
func filterMatrixRows(rows []contracts.MatrixRow, env string) []contracts.MatrixRow {
	if env == "" {
		return rows
	}
	out := make([]contracts.MatrixRow, 0, len(rows))
	for _, row := range rows {
		ed, ok := row.Envs[env]
		if !ok {
			continue
		}
		filtered := row
		filtered.Envs = map[string]contracts.EnvDeploy{env: ed}
		out = append(out, filtered)
	}
	return out
}

var checkDeployCmd = &cobra.Command{
	Use:   "check-deploy <service>",
	Short: "Show what version of a service is deployed where",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		cl, err := newClient()
		if err != nil {
			return err
		}
		// Fetch the fleet matrix and match by name client-side (the backend
		// only exact-matches the full entity ref).
		m, err := cl.Matrix(context.Background(), "", freshFlag)
		if err != nil {
			return err
		}
		matched := filterMatrixByName(m, args[0])
		if len(matched) == 0 {
			return fmt.Errorf("no service matching %q (try the full ref, e.g. component:default/%s)", args[0], args[0])
		}
		rows := filterMatrixRows(matched, checkDeployEnv)
		return render.Output(JSONOutput(), rows, func(w io.Writer) {
			tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
			multi := len(rows) > 1
			if multi {
				fmt.Fprintln(tw, "SERVICE\tENVIRONMENT\tVERSION\tSYNC\tHEALTH")
			} else {
				fmt.Fprintln(tw, "ENVIRONMENT\tVERSION\tSYNC\tHEALTH")
			}
			for _, row := range rows {
				envs := make([]string, 0, len(row.Envs))
				for e := range row.Envs {
					envs = append(envs, e)
				}
				sort.Strings(envs)
				for _, e := range envs {
					d := row.Envs[e]
					if multi {
						fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", row.ServiceName, e, dash(d.Tag), dash(d.SyncStatus), dash(d.HealthStatus))
					} else {
						fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", e, dash(d.Tag), dash(d.SyncStatus), dash(d.HealthStatus))
					}
				}
			}
			tw.Flush()
		})
	},
}

func init() {
	checkDeployCmd.Flags().StringVar(&checkDeployEnv, "env", "", "filter to one environment")
	checkDeployCmd.Flags().BoolVar(&freshFlag, "fresh", false, "bypass server cache")
	RootCmd.AddCommand(checkDeployCmd)
}
