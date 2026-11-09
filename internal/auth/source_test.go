package auth

import (
	"context"
	"testing"
	"time"
)

func TestSourceReturnsValidTokenWithoutRefreshing(t *testing.T) {
	p := newFakePortal(t)
	store := NewStore(t.TempDir())
	store.Save(&Config{PortalURL: p.URL, Token: "cur", RefreshToken: "ref", Expiry: time.Now().Add(time.Hour)})

	tok, err := (&Source{Store: store}).Token(context.Background())
	if err != nil || tok != "cur" {
		t.Fatalf("tok=%q err=%v", tok, err)
	}
	if len(p.forms) != 0 {
		t.Fatalf("refreshed a valid token")
	}
}

func TestSourceRefreshesExpiringTokenAndPersistsRotation(t *testing.T) {
	p := newFakePortal(t)
	p.tokenBody = `{"access_token":"new","refresh_token":"ref2","expires_in":3600}`
	store := NewStore(t.TempDir())
	store.Save(&Config{PortalURL: p.URL, Token: "old", RefreshToken: "ref1", Expiry: time.Now().Add(30 * time.Second)})

	tok, err := (&Source{Store: store}).Token(context.Background())
	if err != nil || tok != "new" {
		t.Fatalf("tok=%q err=%v", tok, err)
	}
	saved, _ := store.Load()
	if saved.Token != "new" || saved.RefreshToken != "ref2" || saved.PortalURL != p.URL {
		t.Fatalf("saved = %+v", saved)
	}
}

func TestSourceLeavesLegacyTokenAlone(t *testing.T) {
	store := NewStore(t.TempDir())
	store.Save(&Config{PortalURL: "http://unused.invalid", Token: "legacy"})

	tok, err := (&Source{Store: store}).Token(context.Background())
	if err != nil || tok != "legacy" {
		t.Fatalf("tok=%q err=%v", tok, err)
	}
}

// A process that waited on the lock must reuse the token another process
// just rotated instead of spending the now-invalid old refresh token.
func TestSourceReloadsAfterLock(t *testing.T) {
	p := newFakePortal(t)
	store := NewStore(t.TempDir())
	store.Save(&Config{PortalURL: p.URL, Token: "old", RefreshToken: "ref1", Expiry: time.Now().Add(-time.Minute)})

	unlock, err := store.Lock()
	if err != nil {
		t.Fatalf("lock: %v", err)
	}
	done := make(chan string)
	go func() {
		tok, _ := (&Source{Store: store}).Token(context.Background())
		done <- tok
	}()
	time.Sleep(100 * time.Millisecond)
	store.Save(&Config{PortalURL: p.URL, Token: "rotated", RefreshToken: "ref2", Expiry: time.Now().Add(time.Hour)})
	unlock()

	if tok := <-done; tok != "rotated" {
		t.Fatalf("tok = %q, want rotated", tok)
	}
	if len(p.forms) != 0 {
		t.Fatalf("refreshed despite another process having rotated")
	}
}
