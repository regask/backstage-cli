package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/regask/backstage-regask-cli/internal/auth"
	"github.com/spf13/cobra"
)

var loginURL string

func openBrowser(target string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", target).Start()
	case "linux":
		return exec.Command("xdg-open", target).Start()
	default:
		fmt.Println("open this URL to sign in:", target)
		return nil
	}
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Sign in to Regask Backstage",
	RunE: func(c *cobra.Command, _ []string) error {
		dir, err := auth.DefaultDir()
		if err != nil {
			return err
		}
		flow := &auth.LoginFlow{Portal: loginURL, OpenBrowser: openBrowser}
		fmt.Println("Opening browser to sign in…")
		cfg, err := flow.Run(context.Background())
		if err != nil {
			return err
		}
		if err := auth.NewStore(dir).Save(cfg); err != nil {
			return err
		}
		fmt.Println("Logged in to", cfg.PortalURL)
		return nil
	},
}

func init() {
	loginCmd.Flags().StringVar(&loginURL, "url", "https://backstage.regask.com", "portal URL")
	RootCmd.AddCommand(loginCmd)
}
