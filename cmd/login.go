package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/regask/backstage-cli/internal/auth"
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
		// Don't hang forever if the user never completes the browser flow.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		fmt.Println("Opening browser to sign in…")
		var cfg *auth.Config
		// Prefer the refreshable OIDC login; fall back to the /cli-auth page
		// (non-refreshable token) while the portal hasn't enabled it.
		if auth.OIDCAvailable(ctx, nil, loginURL) {
			cfg, err = (&auth.OIDCFlow{Portal: loginURL, OpenBrowser: openBrowser}).Run(ctx)
		} else {
			flow := &auth.LoginFlow{
				Portal:      loginURL,
				OpenBrowser: openBrowser,
				Announce: func(code string) {
					fmt.Printf("Confirm this code in your browser: %s\n", code)
				},
			}
			cfg, err = flow.Run(ctx)
		}
		if err != nil {
			return err
		}
		store := auth.NewStore(dir)
		// Don't let an in-flight refresh overwrite the new session.
		unlock, err := store.Lock()
		if err != nil {
			return err
		}
		defer unlock()
		if err := store.Save(cfg); err != nil {
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
