package cmd

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/regask/backstage-cli/internal/render"
	"github.com/spf13/cobra"
)

var findTicketCmd = &cobra.Command{
	Use:   "find-ticket <TICKET> [TICKET...]",
	Short: "Show which environments a ticket is deployed to",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		cl, err := newClient()
		if err != nil {
			return err
		}
		out, err := cl.TicketLookup(context.Background(), args, freshFlag)
		if err != nil {
			return err
		}
		return render.Output(JSONOutput(), out, func(w io.Writer) {
			tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
			fmt.Fprintln(tw, "TICKET\tENVIRONMENTS")
			tickets := make([]string, 0, len(out.Results))
			for k := range out.Results {
				tickets = append(tickets, k)
			}
			sort.Strings(tickets)
			for _, tk := range tickets {
				fmt.Fprintf(tw, "%s\t%s\n", tk, strings.Join(out.Results[tk], ", "))
			}
			tw.Flush()
		})
	},
}

func init() {
	findTicketCmd.Flags().BoolVar(&freshFlag, "fresh", false, "bypass server cache")
	RootCmd.AddCommand(findTicketCmd)
}
