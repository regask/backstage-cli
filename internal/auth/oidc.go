package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"
)

// Backstage auth-backend's built-in OIDC provider. Login and refresh need
// auth.experimentalClientIdMetadataDocuments and auth.experimentalRefreshToken
// enabled; the CLI is the public client the backend itself publishes at
// cliClientPath, which accepts any loopback port.
const (
	authBase      = "/api/auth"
	cliClientPath = authBase + "/.well-known/oauth-client/cli.json"
)

// ErrTokenRejected means the token endpoint refused the grant (expired,
// revoked or already-rotated refresh token); the user must log in again.
var ErrTokenRejected = errors.New("token request rejected")

func defaultHTTP(hc *http.Client) *http.Client {
	if hc != nil {
		return hc
	}
	return &http.Client{Timeout: 30 * time.Second}
}

// OIDCAvailable reports whether the portal can issue refresh tokens to the
// CLI: the CLI client metadata is published and discovery lists the
// refresh_token grant.
func OIDCAvailable(ctx context.Context, hc *http.Client, portal string) bool {
	hc = defaultHTTP(hc)
	if err := getJSON(ctx, hc, portal+cliClientPath, &struct{}{}); err != nil {
		return false
	}
	var disc struct {
		GrantTypes []string `json:"grant_types_supported"`
	}
	if err := getJSON(ctx, hc, portal+authBase+"/.well-known/openid-configuration", &disc); err != nil {
		return false
	}
	return slices.Contains(disc.GrantTypes, "refresh_token")
}

func getJSON(ctx context.Context, hc *http.Client, u string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: %d", u, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// OIDCFlow is an authorization-code + PKCE login that yields a refresh token.
type OIDCFlow struct {
	Portal      string
	OpenBrowser func(url string) error
	HTTP        *http.Client
}

// Run opens the portal's consent page and exchanges the returned code.
func (f *OIDCFlow) Run(ctx context.Context) (*Config, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	defer ln.Close()

	redirect := fmt.Sprintf("http://%s/callback", ln.Addr().String())
	state, err := randomState()
	if err != nil {
		return nil, err
	}
	verifier, err := pkceVerifier()
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])

	codes := make(chan string, 1)
	errc := make(chan error, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		var err error
		switch {
		case q.Get("state") != state:
			err = fmt.Errorf("callback state mismatch")
		case q.Get("error") != "":
			err = fmt.Errorf("login denied: %s %s", q.Get("error"), q.Get("error_description"))
		case q.Get("code") == "":
			err = fmt.Errorf("callback missing code")
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
		case codes <- q.Get("code"):
		default:
		}
	})
	srv := &http.Server{Handler: mux}
	go srv.Serve(ln)
	defer srv.Close()

	clientID := f.Portal + cliClientPath
	q := url.Values{
		"client_id":             {clientID},
		"redirect_uri":          {redirect},
		"response_type":         {"code"},
		"scope":                 {"openid offline_access"},
		"state":                 {state},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
	}
	if err := f.OpenBrowser(f.Portal + authBase + "/v1/authorize?" + q.Encode()); err != nil {
		return nil, err
	}

	var code string
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errc:
		return nil, err
	case code = <-codes:
	}
	cfg, err := postToken(ctx, defaultHTTP(f.HTTP), f.Portal, url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirect},
		"code_verifier": {verifier},
		"client_id":     {clientID},
	})
	if err != nil {
		return nil, err
	}
	if cfg.RefreshToken == "" {
		return nil, fmt.Errorf("portal issued no refresh token (offline_access not granted)")
	}
	return cfg, nil
}

func pkceVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// Refresh trades cfg's refresh token for a new access token. Backstage
// rotates refresh tokens, so the returned Config must replace cfg.
func Refresh(ctx context.Context, hc *http.Client, cfg *Config) (*Config, error) {
	return postToken(ctx, defaultHTTP(hc), cfg.PortalURL, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {cfg.RefreshToken},
	})
}

func postToken(ctx context.Context, hc *http.Client, portal string, form url.Values) (*Config, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, portal+authBase+"/v1/token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w (%d): %s", ErrTokenRejected, resp.StatusCode, strings.TrimSpace(string(b)))
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token endpoint failed (%d)", resp.StatusCode)
	}
	var tr struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, err
	}
	if tr.AccessToken == "" {
		return nil, fmt.Errorf("token endpoint returned no access token")
	}
	return &Config{
		PortalURL:    portal,
		Token:        tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		Expiry:       time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second),
	}, nil
}
