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

// randomHex returns n random bytes hex-encoded, for CSRF states and PKCE
// verifiers. It errors rather than falling back to a predictable value, which
// would defeat both checks.
func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
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

// awaitCallback serves a one-shot loopback /callback, calls open with its
// redirect URI and a fresh CSRF state, and blocks until a callback passes the
// state check and check (or ctx is cancelled). The random state ties the
// callback to this process so another local process can't inject a result.
func awaitCallback(ctx context.Context, open func(redirect, state string) error, check func(url.Values) error) (q url.Values, redirect string, err error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, "", err
	}
	defer ln.Close()

	redirect = fmt.Sprintf("http://%s/callback", ln.Addr().String())
	state, err := randomHex(16)
	if err != nil {
		return nil, "", err
	}
	result := make(chan url.Values, 1)
	errc := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		var err error
		if q.Get("state") != state {
			err = fmt.Errorf("callback state mismatch")
		} else {
			err = check(q)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			select {
			case errc <- err:
			default:
			}
			return
		}
		fmt.Fprintln(w, "Login complete — you can close this tab and return to the terminal.")
		select {
		case result <- q:
		default:
		}
	})

	srv := &http.Server{Handler: mux}
	go srv.Serve(ln)
	defer srv.Close()

	if err := open(redirect, state); err != nil {
		return nil, "", err
	}
	select {
	case <-ctx.Done():
		return nil, "", ctx.Err()
	case err := <-errc:
		return nil, "", err
	case q := <-result:
		return q, redirect, nil
	}
}

// Run opens the browser to the portal's /cli-auth handshake page and blocks
// until the callback delivers the token (or ctx is cancelled).
func (f *LoginFlow) Run(ctx context.Context) (*Config, error) {
	code := pairingCode()
	if f.Announce != nil {
		f.Announce(code)
	}
	q, _, err := awaitCallback(ctx, func(redirect, state string) error {
		return f.OpenBrowser(fmt.Sprintf(
			"%s/cli-auth?redirect_uri=%s&state=%s&code=%s",
			f.Portal, url.QueryEscape(redirect), url.QueryEscape(state), url.QueryEscape(code),
		))
	}, func(q url.Values) error {
		if q.Get("token") == "" {
			return fmt.Errorf("callback missing token")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	cfg := &Config{PortalURL: f.Portal, Token: q.Get("token"), RefreshToken: q.Get("refreshToken")}
	if exp := q.Get("expiry"); exp != "" {
		if t, perr := time.Parse(time.RFC3339, exp); perr == nil {
			cfg.Expiry = t
		}
	}
	return cfg, nil
}
