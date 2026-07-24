package cmd

import (
	"context"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	"github.com/regask/backstage-cli/internal/contracts"
	"github.com/regask/backstage-cli/internal/render"
	"github.com/spf13/cobra"
)

var checkDeployEnv string

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
		m, err := cl.Matrix(context.Background(), args[0], freshFlag)
		if err != nil {
			return err
		}
		rows := filterMatrixRows(m, checkDeployEnv)
		return render.Output(JSONOutput(), rows, func(w io.Writer) {
			tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
			multi := len(rows) > 1
			if multi {
				fmt.Fprintln(tw, "SERVICE\tENVIRONMENT\tVERSION")
			} else {
				fmt.Fprintln(tw, "ENVIRONMENT\tVERSION")
			}
			for _, row := range rows {
				envs := make([]string, 0, len(row.Envs))
				for e := range row.Envs {
					envs = append(envs, e)
				}
				sort.Strings(envs)
				for _, e := range envs {
					if multi {
						fmt.Fprintf(tw, "%s\t%s\t%s\n", row.ServiceName, e, row.Envs[e].Tag)
					} else {
						fmt.Fprintf(tw, "%s\t%s\n", e, row.Envs[e].Tag)
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
