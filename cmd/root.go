package cmd

import "github.com/spf13/cobra"

var jsonOutput bool

var RootCmd = &cobra.Command{
	Use:           "backstage-regask",
	Short:         "Regask Backstage from your terminal",
	Long:          "backstage-regask (bsr) — deploy status, env, secrets, tickets, approvals, and workflows from the shell.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	RootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output machine-readable JSON")
}

// JSONOutput reports whether --json was passed.
func JSONOutput() bool { return jsonOutput }

// Execute runs the root command.
func Execute() error { return RootCmd.Execute() }
