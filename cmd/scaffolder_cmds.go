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
// The scaffolder templates key on the canonical ref (confirm exact form at
// execution alongside the template param names).
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
	releaseEnv      string
	releaseVersion  string
	releaseInclude  []string
	releaseExclude  []string
	cpTag           string
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
		// toEnvironment/services param names: confirm against the promote-code
		// template at execution.
		values := map[string]any{"toEnvironment": promoteToEnv, "services": svcRefs}
		return runTemplate(ctx, cl, "template:default/promote-code", values)
	},
}

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release an environment (creates a release-all approval)",
	RunE: func(c *cobra.Command, _ []string) error {
		if releaseEnv == "" {
			return fmt.Errorf("--env is required")
		}
		if len(releaseInclude) > 0 && len(releaseExclude) > 0 {
			return fmt.Errorf("--include-services and --exclude-services are mutually exclusive")
		}
		cl, err := newClient()
		if err != nil {
			return err
		}
		ctx := context.Background()
		values := map[string]any{"environment": releaseEnv}
		if releaseVersion != "" {
			values["version"] = releaseVersion
		}
		if len(releaseInclude) > 0 {
			refs, err := resolveServices(ctx, cl, releaseInclude)
			if err != nil {
				return err
			}
			values["includeServices"] = refs
		}
		if len(releaseExclude) > 0 {
			refs, err := resolveServices(ctx, cl, releaseExclude)
			if err != nil {
				return err
			}
			values["excludeServices"] = refs
		}
		return runTemplate(ctx, cl, "template:default/release", values)
	},
}

// cherryPickBranches are the only release branches cherry-pick may target.
var cherryPickBranches = []string{"release/preprod", "release/prod"}

var cherryPickCmd = &cobra.Command{
	Use:   "cherry-pick",
	Short: "Cherry-pick a ticket onto a release branch (one PR per repo)",
	RunE: func(c *cobra.Command, _ []string) error {
		if cpTag == "" {
			return fmt.Errorf("--tag is required (a ticket reference, e.g. REG-12345)")
		}
		switch cpBranch {
		case "release/preprod", "release/prod":
		default:
			return fmt.Errorf("--branch must be one of: %s", strings.Join(cherryPickBranches, ", "))
		}
		cl, err := newClient()
		if err != nil {
			return err
		}
		return runTemplate(context.Background(), cl, "template:default/cherry-pick", map[string]any{"tag": cpTag, "branch": cpBranch})
	},
}

func init() {
	promoteCmd.Flags().StringVar(&promoteToEnv, "to-env", "", "target environment to promote to, e.g. staging (required)")
	promoteCmd.Flags().StringSliceVar(&promoteServices, "service", nil, "service(s) to promote — repeatable or comma-separated (required)")
	releaseCmd.Flags().StringVar(&releaseEnv, "env", "", "environment to release (required)")
	releaseCmd.Flags().StringVar(&releaseVersion, "version", "", "override shared-actions release version (prod only)")
	releaseCmd.Flags().StringSliceVar(&releaseInclude, "include-services", nil, "only release these services (comma-separated; mutually exclusive with --exclude-services)")
	releaseCmd.Flags().StringSliceVar(&releaseExclude, "exclude-services", nil, "release all except these services (comma-separated; mutually exclusive with --include-services)")
	cherryPickCmd.Flags().StringVar(&cpTag, "tag", "", "ticket reference to cherry-pick, e.g. REG-12345 (required)")
	cherryPickCmd.Flags().StringVar(&cpBranch, "branch", "", "target release branch: release/preprod or release/prod (required)")
	RootCmd.AddCommand(promoteCmd, releaseCmd, cherryPickCmd)
}
