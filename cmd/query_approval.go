package cmd

import (
	"context"
	"io"

	"github.com/regask/backstage-cli/internal/client"
	"github.com/regask/backstage-cli/internal/render"
	"github.com/spf13/cobra"
)

var queryApprovalCmd = &cobra.Command{
	Use:   "query-approval <backstage-link-or-id>",
	Short: "Show approval details, including the release link (read-only)",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		cl, err := newClient()
		if err != nil {
			return err
		}
		id, err := client.ParseApprovalID(args[0])
		if err != nil {
			return err
		}
		req, err := cl.GetApproval(context.Background(), id)
		if err != nil {
			return err
		}
		return render.Output(JSONOutput(), req, func(w io.Writer) {
			printApprovalDetail(w, cl.BaseURL(), req)
		})
	},
}

func init() { RootCmd.AddCommand(queryApprovalCmd) }
