package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/regask/backstage-cli/internal/scaffolder"
	"github.com/spf13/cobra"
)

// runTemplate launches a template and streams its log; non-zero exit on failure.
func runTemplate(templateRef string, values map[string]any) error {
	cl, err := newClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
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
		// toEnvironment/services param names: confirm against the promote-code
		// template at execution.
		values := map[string]any{"toEnvironment": promoteToEnv, "services": promoteServices}
		return runTemplate("template:default/promote-code", values)
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
		values := map[string]any{"environment": releaseEnv}
		if releaseVersion != "" {
			values["version"] = releaseVersion
		}
		if len(releaseInclude) > 0 {
			values["includeServices"] = releaseInclude
		}
		if len(releaseExclude) > 0 {
			values["excludeServices"] = releaseExclude
		}
		return runTemplate("template:default/release", values)
	},
}

var cherryPickCmd = &cobra.Command{
	Use:   "cherry-pick",
	Short: "Cherry-pick a tag onto a release branch (one PR per repo)",
	RunE: func(c *cobra.Command, _ []string) error {
		if cpTag == "" || cpBranch == "" {
			return fmt.Errorf("--tag and --branch are required")
		}
		return runTemplate("template:default/cherry-pick", map[string]any{"tag": cpTag, "branch": cpBranch})
	},
}

func init() {
	promoteCmd.Flags().StringVar(&promoteToEnv, "to-env", "", "target environment to promote to, e.g. staging (required)")
	promoteCmd.Flags().StringSliceVar(&promoteServices, "service", nil, "service(s) to promote — repeatable or comma-separated (required)")
	releaseCmd.Flags().StringVar(&releaseEnv, "env", "", "environment to release (required)")
	releaseCmd.Flags().StringVar(&releaseVersion, "version", "", "override shared-actions release version (prod only)")
	releaseCmd.Flags().StringSliceVar(&releaseInclude, "include-services", nil, "only release these services (comma-separated; mutually exclusive with --exclude-services)")
	releaseCmd.Flags().StringSliceVar(&releaseExclude, "exclude-services", nil, "release all except these services (comma-separated; mutually exclusive with --include-services)")
	cherryPickCmd.Flags().StringVar(&cpTag, "tag", "", "tag to cherry-pick (required)")
	cherryPickCmd.Flags().StringVar(&cpBranch, "branch", "", "target release branch (required)")
	RootCmd.AddCommand(promoteCmd, releaseCmd, cherryPickCmd)
}
