package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/regask/backstage-cli/internal/contracts"
)

// printApprovalDetail renders an approval's human view, including the release
// link (resultUrl) and a backlink to the originating scaffolder task.
func printApprovalDetail(w io.Writer, portalURL string, r contracts.ApprovalRequest) {
	fmt.Fprintf(w, "%s  [%s]  status=%s\n", r.Title, r.Kind, r.Status)
	if r.Requester != "" {
		fmt.Fprintf(w, "requested by %s\n", r.Requester)
	}
	if r.Summary != "" {
		fmt.Fprintln(w, "\n"+r.Summary)
	}
	if r.ResultURL != "" {
		fmt.Fprintf(w, "\nrelease link: %s\n", r.ResultURL)
	} else if drafts := r.DraftReleaseURLs(); len(drafts) > 0 {
		if len(drafts) == 1 {
			fmt.Fprintf(w, "\ndraft release: %s\n", drafts[0])
		} else {
			fmt.Fprintln(w, "\ndraft releases:")
			for _, d := range drafts {
				fmt.Fprintf(w, "  %s\n", d)
			}
		}
	}
	if task := r.TaskID(); task != "" {
		base := strings.TrimRight(portalURL, "/")
		fmt.Fprintf(w, "task:         %s/scaffolder/tasks/%s\n", base, task)
	}
	if r.Error != "" {
		fmt.Fprintf(w, "\nerror: %s\n", r.Error)
	}
}
