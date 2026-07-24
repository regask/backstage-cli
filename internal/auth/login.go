package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type LoginFlow struct {
	Portal      string
	OpenBrowser func(url string) error
	// Announce, if set, is called with the pairing code before the browser
	// opens, so the caller can print it for the user to match on the consent
	// screen. Keeps this package free of terminal output.
	Announce func(code string)
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

// randomState returns a CSRF token tying the browser handoff to this process.
func randomState() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// rand.Read never fails on supported platforms; fall back defensively.
		return "cli-state"
	}
	return hex.EncodeToString(b)
}

// pairingCode returns a short human-matchable code the user compares between
// their terminal and the browser consent screen.
func pairingCode() string {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "CLICODE"
	}
	return strings.ToUpper(hex.EncodeToString(b))
}

// Run starts an ephemeral loopback server, opens the browser to the portal's
// /cli-auth handshake page with the loopback as redirect_uri, and blocks until
// the callback delivers the token (or ctx is cancelled). A random state ties
// the callback to this process so another local process can't inject a token.
func (f *LoginFlow) Run(ctx context.Context) (*Config, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	defer ln.Close()

	redirect := fmt.Sprintf("http://%s/callback", ln.Addr().String())
	state := randomState()
	code := pairingCode()
	result := make(chan *Config, 1)
	errc := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("state") != state {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			errc <- fmt.Errorf("callback state mismatch")
			return
		}
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

	if f.Announce != nil {
		f.Announce(code)
	}
	startURL := fmt.Sprintf(
		"%s/cli-auth?redirect_uri=%s&state=%s&code=%s",
		f.Portal, url.QueryEscape(redirect), url.QueryEscape(state), url.QueryEscape(code),
	)
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
