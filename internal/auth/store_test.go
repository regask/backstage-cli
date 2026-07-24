package auth

import (
	"os"
	"testing"
)

func TestStoreRoundTripAndPerms(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)

	if err := s.Save(&Config{PortalURL: "https://portal.example", Token: "tok"}); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := s.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.Token != "tok" || got.PortalURL != "https://portal.example" {
		t.Fatalf("round-trip mismatch: %+v", got)
	}

	info, err := os.Stat(dir + "/config.json")
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("want file mode 0600, got %o", info.Mode().Perm())
	}
}

func TestClearIsIdempotent(t *testing.T) {
	s := NewStore(t.TempDir())
	if err := s.Clear(); err != nil {
		t.Fatalf("clear on empty should be nil, got %v", err)
	}
}
