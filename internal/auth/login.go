package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

type LoginFlow struct {
	Portal      string
	OpenBrowser func(url string) error
}

// extractRedirectURI pulls the redirect_uri query param out of a start URL
// (exported-for-test via the unexported helper used by the test in-package).
func extractRedirectURI(startURL string) string {
	u, err := url.Parse(startURL)
	if err != nil {
		return ""
	}
	return u.Query().Get("redirect_uri")
}

// Run starts an ephemeral loopback server, opens the browser to the portal's
// CLI sign-in with the loopback as redirect_uri, and blocks until the callback
// delivers the token (or ctx is cancelled).
func (f *LoginFlow) Run(ctx context.Context) (*Config, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	defer ln.Close()

	redirect := fmt.Sprintf("http://%s/callback", ln.Addr().String())
	result := make(chan *Config, 1)
	errc := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		tok := q.Get("token")
		if tok == "" {
			http.Error(w, "missing token", http.StatusBadRequest)
			errc <- fmt.Errorf("callback missing token")
			return
		}
		cfg := &Config{PortalURL: f.Portal, Token: tok, RefreshToken: q.Get("refreshToken")}
		if exp := q.Get("expiry"); exp != "" {
			if t, perr := time.Parse(time.RFC3339, exp); perr == nil {
				cfg.Expiry = t
			}
		}
		fmt.Fprintln(w, "Login complete — you can close this tab and return to the terminal.")
		result <- cfg
	})

	srv := &http.Server{Handler: mux}
	go srv.Serve(ln)
	defer srv.Close()

	// TODO(execution): confirm the real portal start path against the backstage
	// auth config. Shape: <portal>/cli-login?redirect_uri=<loopback>/callback
	startURL := fmt.Sprintf("%s/cli-login?redirect_uri=%s", f.Portal, url.QueryEscape(redirect))
	if err := f.OpenBrowser(startURL); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errc:
		return nil, err
	case cfg := <-result:
		return cfg, nil
	}
}
