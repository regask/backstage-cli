package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/regask/backstage-cli/internal/auth"
	"github.com/regask/backstage-cli/internal/client"
)

var freshFlag bool

// resolveServiceRef turns a bare service name (alert-service) into its full
// entity ref (component:default/alert-service) by matching against the fleet
// matrix. Backend endpoints (matrix filter, overlays) exact-match the full
// ref, so a bare name must be resolved first. An input already containing "/"
// is treated as a full ref and returned as-is.
func resolveServiceRef(ctx context.Context, cl *client.Client, nameOrRef string, fresh bool) (string, error) {
	if strings.Contains(nameOrRef, "/") {
		return nameOrRef, nil
	}
	rows, err := cl.Matrix(ctx, "", fresh)
	if err != nil {
		return "", err
	}
	for _, row := range rows {
		if matchesService(row, nameOrRef) {
			return row.ServiceRef, nil
		}
	}
	return "", fmt.Errorf("no service matching %q (try the full ref, e.g. component:default/%s)", nameOrRef, nameOrRef)
}

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
