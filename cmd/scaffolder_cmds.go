package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/regask/backstage-cli/internal/client"
	"github.com/regask/backstage-cli/internal/scaffolder"
	"github.com/spf13/cobra"
)

// runTemplate launches a template and streams its log; non-zero exit on failure.
func runTemplate(ctx context.Context, cl *client.Client, templateRef string, values map[string]any) error {
	id, err := scaffolder.Launch(ctx, cl, templateRef, values)
	if err != nil {
		return err
	}
	fmt.Println("Launched task", id)
	status, err := scaffolder.Stream(ctx, cl, id, os.Stdout)
	if err != nil {
		return err
	}
	if status != "completed" {
		return fmt.Errorf("task %s finished with status %q", id, status)
	}
	fmt.Printf("Task %s completed. Check `bsr` approvals or the portal for any created approval.\n", id)
	return nil
}

// resolveServices maps bare service names to full entity refs in one matrix
// fetch, so promote/release accept `alert-service` like check-deploy does.
func resolveServices(ctx context.Context, cl *client.Client, names []string) ([]string, error) {
	rows, err := cl.Matrix(ctx, "", false)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(names))
	for _, n := range names {
		if strings.Contains(n, "/") {
			out = append(out, n) // already a ref
			continue
		}
		ref := ""
		for _, row := range rows {
			if matchesService(row, n) {
				ref = row.ServiceRef
				break
			}
		}
		if ref == "" {
			return nil, fmt.Errorf("no service matching %q (try the full ref, e.g. component:default/%s)", n, n)
		}
		out = append(out, ref)
	}
	return out, nil
}

var (
	promoteToEnv    string
	promoteServices []string
	promoteTag      string
	releaseEnv      string
	releaseVersion  int
	releaseExclude  []string
	cpTickets       []string
	cpBranch        string
)

var promoteCmd = &cobra.Command{
	Use:   "promote",
	Short: "Promote code to an environment (creates a draft release + release-publish approval)",
	RunE: func(c *cobra.Command, _ []string) error {
		if promoteToEnv == "" {
			return fmt.Errorf("--to-env is required")
		}
		if len(promoteServices) == 0 {
			return fmt.Errorf("--service is required (one or more)")
		}
		cl, err := newClient()
		if err != nil {
			return err
		}
		ctx := context.Background()
		svcRefs, err := resolveServices(ctx, cl, promoteServices)
		if err != nil {
			return err
		}
		values, err := scaffolder.PromoteValues(ctx, cl, promoteToEnv, svcRefs, promoteTag)
		if err != nil {
			return err
		}
		return runTemplate(ctx, cl, scaffolder.PromoteTemplate, values)
	},
}

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release an environment (creates a release-all approval)",
	RunE: func(c *cobra.Command, _ []string) error {
		if releaseEnv == "" {
			return fmt.Errorf("--env is required")
		}
		cl, err := newClient()
		if err != nil {
			return err
		}
		ctx := context.Background()
		var refs []string
		if len(releaseExclude) > 0 {
			if refs, err = resolveServices(ctx, cl, releaseExclude); err != nil {
				return err
			}
		}
		values, err := scaffolder.ReleaseValues(ctx, cl, releaseEnv, refs, releaseVersion)
		if err != nil {
			return err
		}
		return runTemplate(ctx, cl, scaffolder.ReleaseTemplate, values)
	},
}

var cherryPickCmd = &cobra.Command{
	Use:   "cherry-pick",
	Short: "Cherry-pick tickets onto a release branch (one PR per repo)",
	RunE: func(c *cobra.Command, _ []string) error {
		values, err := scaffolder.CherryPickValues(cpTickets, cpBranch)
		if err != nil {
			return err
		}
		cl, err := newClient()
		if err != nil {
			return err
		}
		return runTemplate(context.Background(), cl, scaffolder.CherryPickTemplate, values)
	},
}

func init() {
	promoteCmd.Flags().StringVar(&promoteToEnv, "to-env", "", "target environment: "+strings.Join(scaffolder.PromoteTargets, ", ")+" (required)")
	promoteCmd.Flags().StringSliceVar(&promoteServices, "service", nil, "service(s) to promote — repeatable or comma-separated (required)")
	promoteCmd.Flags().StringVar(&promoteTag, "tag", "", "source tag to promote (default: latest at the appropriate level)")
	releaseCmd.Flags().StringVar(&releaseEnv, "env", "", "environment to release: "+strings.Join(scaffolder.ReleaseTargets, ", ")+" (required)")
	releaseCmd.Flags().IntVar(&releaseVersion, "version", 0, "bump the shared-actions release version (production only)")
	releaseCmd.Flags().StringSliceVar(&releaseExclude, "exclude-services", nil, "release all except these services (comma-separated)")
	cherryPickCmd.Flags().StringSliceVar(&cpTickets, "tag", nil, "ticket(s) to cherry-pick, e.g. REG-12345 — repeatable or comma-separated (required)")
	cherryPickCmd.Flags().StringVar(&cpBranch, "branch", "", "target release branch: "+strings.Join(scaffolder.CherryPickBranches, ", ")+" (required)")
	RootCmd.AddCommand(promoteCmd, releaseCmd, cherryPickCmd)
}
