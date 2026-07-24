package auth

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"
)

// startParams pulls redirect_uri + state out of the start URL the flow opens.
func startParams(t *testing.T, startURL string) (redirect, state string) {
	t.Helper()
	u, err := url.Parse(startURL)
	if err != nil {
		t.Fatalf("parse start url: %v", err)
	}
	return u.Query().Get("redirect_uri"), u.Query().Get("state")
}

func TestLoginFlowCapturesTokenFromCallback(t *testing.T) {
	flow := &LoginFlow{
		Portal: "https://portal.example",
		OpenBrowser: func(startURL string) error {
			// Simulate the /cli-auth page redirecting to the loopback callback
			// with the token and the echoed CSRF state.
			go func() {
				redirect, state := startParams(t, startURL)
				http.Get(redirect + "?token=abc&refreshToken=ref&state=" +
					url.QueryEscape(state) + "&expiry=2030-01-01T00:00:00Z")
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

func TestLoginFlowRejectsStateMismatch(t *testing.T) {
	flow := &LoginFlow{
		Portal: "https://portal.example",
		OpenBrowser: func(startURL string) error {
			go func() {
				redirect, _ := startParams(t, startURL)
				// Wrong state — a foreign process must not be able to inject a token.
				http.Get(redirect + "?token=abc&state=WRONG")
			}()
			return nil
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := flow.Run(ctx); err == nil {
		t.Fatalf("expected state-mismatch error, got nil")
	}
}
