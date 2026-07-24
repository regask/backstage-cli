package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/regask/backstage-cli/internal/render"
	"github.com/spf13/cobra"
)

// displayServiceRef strips the "component:default/" prefix from a service
// entity ref for display.
func displayServiceRef(ref string) string {
	return strings.TrimPrefix(ref, "component:default/")
}

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
			fmt.Fprintln(tw, "SERVICE\tENVIRONMENTS\tTICKETS")
			for _, res := range out.Services {
				seen := make(map[string]bool, len(res.Commits))
				tickets := make([]string, 0, len(res.Commits))
				for _, cm := range res.Commits {
					if cm.Ticket == "" || seen[cm.Ticket] {
						continue
					}
					seen[cm.Ticket] = true
					tickets = append(tickets, cm.Ticket)
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\n", displayServiceRef(res.ServiceRef), strings.Join(res.DeployedEnvs, ", "), strings.Join(tickets, ", "))
			}
			tw.Flush()
			if len(out.NotFound) > 0 {
				fmt.Fprintf(w, "not found: %s\n", strings.Join(out.NotFound, ", "))
			}
		})
	},
}

func init() {
	findTicketCmd.Flags().BoolVar(&freshFlag, "fresh", false, "bypass server cache")
	RootCmd.AddCommand(findTicketCmd)
}
