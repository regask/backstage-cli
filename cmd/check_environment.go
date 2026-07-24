package cmd

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/regask/backstage-cli/internal/contracts"
	"github.com/regask/backstage-cli/internal/render"
	"github.com/spf13/cobra"
)

var checkEnvEnv string

var checkEnvironmentCmd = &cobra.Command{
	Use:   "check-environment <service>",
	Short: "Show the effective environment variables for a service in an environment",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		if checkEnvEnv == "" {
			return fmt.Errorf("--env is required")
		}
		cl, err := newClient()
		if err != nil {
			return err
		}
		ov, err := cl.Overlays(context.Background(), args[0], freshFlag)
		if err != nil {
			return err
		}
		effective := contracts.MergeMaps(
			contracts.ParseEnvFile(ov.BaseEnv),
			contracts.ParseEnvFile(ov.EnvOverlays[checkEnvEnv]),
		)
		return render.Output(JSONOutput(), effective, func(w io.Writer) {
			tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
			fmt.Fprintln(tw, "KEY\tVALUE")
			for _, k := range contracts.SortedKeys(effective) {
				fmt.Fprintf(tw, "%s\t%s\n", k, effective[k])
			}
			tw.Flush()
		})
	},
}

func init() {
	checkEnvironmentCmd.Flags().StringVar(&checkEnvEnv, "env", "", "environment (required)")
	checkEnvironmentCmd.Flags().BoolVar(&freshFlag, "fresh", false, "bypass server cache")
	RootCmd.AddCommand(checkEnvironmentCmd)
}
