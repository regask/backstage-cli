package cmd

import (
	"context"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	"github.com/regask/backstage-regask-cli/internal/render"
	"github.com/spf13/cobra"
)

var checkDeployEnv string

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
		return render.Output(JSONOutput(), m, func(w io.Writer) {
			tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
			fmt.Fprintln(tw, "ENVIRONMENT\tVERSION")
			for _, row := range m.Rows {
				envs := make([]string, 0, len(row.Environments))
				for e := range row.Environments {
					if checkDeployEnv == "" || e == checkDeployEnv {
						envs = append(envs, e)
					}
				}
				sort.Strings(envs)
				for _, e := range envs {
					fmt.Fprintf(tw, "%s\t%s\n", e, row.Environments[e])
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
