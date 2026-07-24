package cmd

import (
	"fmt"

	"github.com/regask/backstage-cli/internal/auth"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear the stored credential",
	RunE: func(c *cobra.Command, _ []string) error {
		dir, err := auth.DefaultDir()
		if err != nil {
			return err
		}
		if err := auth.NewStore(dir).Clear(); err != nil {
			return err
		}
		fmt.Println("Logged out.")
		return nil
	},
}

func init() { RootCmd.AddCommand(logoutCmd) }
