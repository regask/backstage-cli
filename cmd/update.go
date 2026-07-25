package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// lookPath / runStreaming are overridable in tests.
var lookPath = exec.LookPath

func runStreaming(name string, args ...string) error {
	c := exec.Command(name, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	return c.Run()
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update backstage-regask to the latest release (via Homebrew)",
	Long: "Updates backstage-regask to the latest tap release using Homebrew " +
		"(`brew update` then `brew upgrade`). If you installed another way, " +
		"update via that method instead.",
	RunE: func(c *cobra.Command, _ []string) error {
		if _, err := lookPath("brew"); err != nil {
			return fmt.Errorf("brew not found — this command updates via Homebrew; " +
				"reinstall from the tap (brew install regask/tap/backstage-regask) or use your install method")
		}
		fmt.Println("Refreshing Homebrew…")
		if err := runStreaming("brew", "update"); err != nil {
			return fmt.Errorf("brew update failed: %w", err)
		}
		fmt.Println("Upgrading backstage-regask…")
		if err := runStreaming("brew", "upgrade", "backstage-regask"); err != nil {
			return fmt.Errorf("brew upgrade failed: %w", err)
		}
		fmt.Println("Done — run `bsr --version` to confirm.")
		return nil
	},
}

func init() { RootCmd.AddCommand(updateCmd) }
