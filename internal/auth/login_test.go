package auth

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestLoginFlowCapturesTokenFromCallback(t *testing.T) {
	flow := &LoginFlow{
		Portal: "https://portal.example",
		OpenBrowser: func(startURL string) error {
			// Simulate the browser completing auth and redirecting to the
			// loopback callback with the token. startURL carries redirect_uri.
			go func() {
				redirect := extractRedirectURI(startURL)
				http.Get(redirect + "?token=abc&refreshToken=ref&expiry=2030-01-01T00:00:00Z")
			}()
			return nil
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := flow.Run(ctx)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if cfg.Token != "abc" || cfg.RefreshToken != "ref" {
		t.Fatalf("captured %+v", cfg)
	}
	if cfg.PortalURL != "https://portal.example" {
		t.Fatalf("portal = %q", cfg.PortalURL)
	}
}
