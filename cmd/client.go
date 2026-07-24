package cmd

import (
	"github.com/regask/backstage-regask-cli/internal/auth"
	"github.com/regask/backstage-regask-cli/internal/client"
)

var freshFlag bool

// tokenFn loads the stored token or returns ErrUnauthorized.
func tokenFn() (string, error) {
	dir, err := auth.DefaultDir()
	if err != nil {
		return "", err
	}
	cfg, err := auth.NewStore(dir).Load()
	if err != nil || cfg.Token == "" {
		return "", client.ErrUnauthorized
	}
	return cfg.Token, nil
}

// newClient builds a client pointed at the stored portal URL.
func newClient() (*client.Client, error) {
	dir, err := auth.DefaultDir()
	if err != nil {
		return nil, err
	}
	cfg, err := auth.NewStore(dir).Load()
	if err != nil || cfg.PortalURL == "" {
		return nil, client.ErrUnauthorized
	}
	return client.New(cfg.PortalURL, tokenFn, nil), nil
}
