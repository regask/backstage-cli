package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/regask/backstage-cli/internal/auth"
	"github.com/regask/backstage-cli/internal/client"
	"github.com/regask/backstage-cli/internal/render"
	"github.com/spf13/cobra"
)

// subjectFromToken returns the JWT `sub` claim (the Backstage user entity ref).
func subjectFromToken(tok string) (string, error) {
	parts := strings.Split(tok, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("not a JWT")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}
	var claims struct {
		Sub string `json:"sub"`
	}
	if err := json.Unmarshal(raw, &claims); err != nil {
		return "", err
	}
	return claims.Sub, nil
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show the signed-in user",
	RunE: func(c *cobra.Command, _ []string) error {
		dir, err := auth.DefaultDir()
		if err != nil {
			return err
		}
		cfg, err := auth.NewStore(dir).Load()
		if err != nil || cfg.Token == "" {
			return client.ErrUnauthorized
		}
		sub, err := subjectFromToken(cfg.Token)
		if err != nil {
			return err
		}
		out := struct {
			Subject string `json:"subject"`
		}{Subject: sub}
		return render.Output(JSONOutput(), out, func(w io.Writer) {
			fmt.Fprintln(w, out.Subject)
		})
	},
}

func init() { RootCmd.AddCommand(whoamiCmd) }
