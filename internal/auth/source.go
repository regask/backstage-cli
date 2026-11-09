package auth

import (
	"context"
	"net/http"
	"time"
)

// expirySkew refreshes slightly early so a token doesn't expire mid-request.
const expirySkew = time.Minute

// Source hands out the stored access token, refreshing it first when it is
// about to expire and a refresh token is stored.
type Source struct {
	Store *Store
	HTTP  *http.Client
	Now   func() time.Time
}

func (s *Source) Token(ctx context.Context) (string, error) {
	cfg, err := s.Store.Load()
	if err != nil {
		return "", err
	}
	if !s.needsRefresh(cfg) {
		return cfg.Token, nil
	}
	unlock, err := s.Store.Lock()
	if err != nil {
		return "", err
	}
	defer unlock()
	// Another bsr process may have rotated the token while we waited; its
	// predecessor is now single-use spent and would be rejected.
	if cfg, err = s.Store.Load(); err != nil {
		return "", err
	}
	if !s.needsRefresh(cfg) {
		return cfg.Token, nil
	}
	next, err := Refresh(ctx, s.HTTP, cfg)
	if err != nil {
		return "", err
	}
	if err := s.Store.Save(next); err != nil {
		return "", err
	}
	return next.Token, nil
}

func (s *Source) needsRefresh(c *Config) bool {
	now := time.Now
	if s.Now != nil {
		now = s.Now
	}
	return c.RefreshToken != "" && !c.Expiry.IsZero() && !now().Before(c.Expiry.Add(-expirySkew))
}
