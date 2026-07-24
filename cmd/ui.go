package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/regask/backstage-cli/internal/auth"
	"github.com/regask/backstage-cli/internal/tui"
	"github.com/spf13/cobra"
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Open the interactive terminal UI (k9s-style)",
	RunE: func(c *cobra.Command, _ []string) error {
		cl, err := newClient()
		if err != nil {
			return err
		}
		dir, _ := auth.DefaultDir()
		cfg, _ := auth.NewStore(dir).Load()
		user := ""
		if cfg != nil {
			if sub, e := subjectFromToken(cfg.Token); e == nil {
				user = sub
			}
		}
		portal := ""
		if cfg != nil {
			portal = cfg.PortalURL
		}
		p := tea.NewProgram(tui.NewApp(cl, portal, user), tea.WithAltScreen())
		_, err = p.Run()
		return err
	},
}

func init() { RootCmd.AddCommand(uiCmd) }
