package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// fakePortal serves the auth-backend OIDC endpoints the CLI uses. The token
// handler records each form it receives.
type fakePortal struct {
	*httptest.Server
	cimd       bool
	grantTypes []string
	tokenCode  int
	tokenBody  string
	forms      []url.Values
}

func newFakePortal(t *testing.T) *fakePortal {
	p := &fakePortal{cimd: true, grantTypes: []string{"authorization_code", "refresh_token"}, tokenCode: 200,
		tokenBody: `{"access_token":"acc","refresh_token":"ref","expires_in":3600,"token_type":"Bearer"}`}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/.well-known/oauth-client/cli.json", func(w http.ResponseWriter, _ *http.Request) {
		if !p.cimd {
			http.Error(w, `{"error":"not_found"}`, http.StatusNotFound)
			return
		}
		w.Write([]byte(`{"client_name":"Backstage CLI"}`))
	})
	mux.HandleFunc("/api/auth/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"grant_types_supported": p.grantTypes})
	})
	mux.HandleFunc("/api/auth/v1/token", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		p.forms = append(p.forms, r.PostForm)
		w.WriteHeader(p.tokenCode)
		w.Write([]byte(p.tokenBody))
	})
	p.Server = httptest.NewServer(mux)
	t.Cleanup(p.Close)
	return p
}

func TestOIDCAvailable(t *testing.T) {
	p := newFakePortal(t)
	ctx := context.Background()
	if !OIDCAvailable(ctx, nil, p.URL) {
		t.Fatalf("want available")
	}
	p.grantTypes = []string{"authorization_code"}
	if OIDCAvailable(ctx, nil, p.URL) {
		t.Fatalf("no refresh_token grant: want unavailable")
	}
	p.grantTypes, p.cimd = []string{"refresh_token"}, false
	if OIDCAvailable(ctx, nil, p.URL) {
		t.Fatalf("CIMD disabled: want unavailable")
	}
}

func TestOIDCFlowExchangesCodeWithPKCE(t *testing.T) {
	p := newFakePortal(t)
	var authorize url.Values
	flow := &OIDCFlow{Portal: p.URL, OpenBrowser: func(u string) error {
		parsed, err := url.Parse(u)
		if err != nil {
			return err
		}
		authorize = parsed.Query()
		// Simulate the consent page redirecting back with a code.
		go http.Get(authorize.Get("redirect_uri") + "?code=c1&state=" + url.QueryEscape(authorize.Get("state")))
		return nil
	}}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := flow.Run(ctx)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if cfg.Token != "acc" || cfg.RefreshToken != "ref" || cfg.PortalURL != p.URL {
		t.Fatalf("cfg = %+v", cfg)
	}
	if until := time.Until(cfg.Expiry); until < 59*time.Minute || until > time.Hour {
		t.Fatalf("expiry in %v, want ~1h", until)
	}
	if authorize.Get("client_id") != p.URL+"/api/auth/.well-known/oauth-client/cli.json" ||
		authorize.Get("scope") != "openid offline_access" || authorize.Get("code_challenge_method") != "S256" {
		t.Fatalf("authorize params = %v", authorize)
	}
	form := p.forms[0]
	sum := sha256.Sum256([]byte(form.Get("code_verifier")))
	if base64.RawURLEncoding.EncodeToString(sum[:]) != authorize.Get("code_challenge") {
		t.Fatalf("code_verifier does not match code_challenge")
	}
	if form.Get("grant_type") != "authorization_code" || form.Get("code") != "c1" ||
		form.Get("redirect_uri") != authorize.Get("redirect_uri") {
		t.Fatalf("token form = %v", form)
	}
}

func TestOIDCFlowRejectsStateMismatch(t *testing.T) {
	p := newFakePortal(t)
	flow := &OIDCFlow{Portal: p.URL, OpenBrowser: func(u string) error {
		parsed, _ := url.Parse(u)
		go http.Get(parsed.Query().Get("redirect_uri") + "?code=c1&state=WRONG")
		return nil
	}}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := flow.Run(ctx); err == nil {
		t.Fatalf("expected state-mismatch error")
	}
	if len(p.forms) != 0 {
		t.Fatalf("exchanged a code despite state mismatch")
	}
}

func TestRefreshRejectedMapsToSentinel(t *testing.T) {
	p := newFakePortal(t)
	p.tokenCode, p.tokenBody = 400, `{"error":"invalid_grant"}`
	_, err := Refresh(context.Background(), nil, &Config{PortalURL: p.URL, RefreshToken: "old"})
	if !errors.Is(err, ErrTokenRejected) {
		t.Fatalf("want ErrTokenRejected, got %v", err)
	}
	if p.forms[0].Get("grant_type") != "refresh_token" || p.forms[0].Get("refresh_token") != "old" {
		t.Fatalf("refresh form = %v", p.forms[0])
	}
}

// Only 400/401 mean the session is dead; a throttled or flaky endpoint must
// not force a re-login.
func TestRefreshTransientErrorIsNotRejection(t *testing.T) {
	p := newFakePortal(t)
	p.tokenCode, p.tokenBody = 429, `slow down`
	_, err := Refresh(context.Background(), nil, &Config{PortalURL: p.URL, RefreshToken: "old"})
	if err == nil || errors.Is(err, ErrTokenRejected) {
		t.Fatalf("want non-rejection error, got %v", err)
	}
}

func TestRefreshKeepsTokenWhenNotRotated(t *testing.T) {
	p := newFakePortal(t)
	p.tokenBody = `{"access_token":"new","expires_in":3600}`
	next, err := Refresh(context.Background(), nil, &Config{PortalURL: p.URL, RefreshToken: "keep"})
	if err != nil || next.RefreshToken != "keep" || next.Token != "new" {
		t.Fatalf("next=%+v err=%v", next, err)
	}
}
