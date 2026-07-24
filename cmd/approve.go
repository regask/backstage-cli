package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/regask/backstage-cli/internal/client"
	"github.com/spf13/cobra"
)

var approveReject bool

var approveCmd = &cobra.Command{
	Use:   "approve <backstage-link-or-id>",
	Short: "Approve (or --reject) an approval request",
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
		ctx := context.Background()
		req, err := cl.GetApproval(ctx, id)
		if err != nil {
			return err
		}
		printApprovalDetail(os.Stdout, cl.BaseURL(), req)
		verb := "approve"
		if approveReject {
			verb = "reject"
		}
		fmt.Printf("\n%s this request? [y/N] ", verb)
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(line)) != "y" {
			return fmt.Errorf("aborted")
		}
		if err := cl.DecideApproval(ctx, id, !approveReject); err != nil {
			return err
		}
		fmt.Printf("Request %s %sd.\n", id, verb)
		return nil
	},
}

func init() {
	approveCmd.Flags().BoolVar(&approveReject, "reject", false, "reject instead of approve")
	RootCmd.AddCommand(approveCmd)
}
