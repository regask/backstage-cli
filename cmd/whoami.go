package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/regask/backstage-cli/internal/render"
	"github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show the signed-in user",
	RunE: func(c *cobra.Command, _ []string) error {
		cl, err := newClient()
		if err != nil {
			return err
		}
		var out struct {
			Identity struct {
				UserEntityRef string `json:"userEntityRef"`
			} `json:"identity"`
		}
		// TODO(execution): confirm identity endpoint; Backstage default is
		// /api/auth/.../refresh or a user-info route.
		if err := cl.GetJSON(context.Background(), "/auth/v1/userinfo", nil, false, &out); err != nil {
			return err
		}
		return render.Output(JSONOutput(), out, func(w io.Writer) {
			fmt.Fprintln(w, out.Identity.UserEntityRef)
		})
	},
}

func init() { RootCmd.AddCommand(whoamiCmd) }
