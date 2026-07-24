package auth

import (
	"os"
	"path/filepath"
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

func TestSaveTightensExistingLoosePerms(t *testing.T) {
	base := filepath.Join(t.TempDir(), "cfg")
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(base, "config.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	s := NewStore(base)
	if err := s.Save(&Config{Token: "x"}); err != nil {
		t.Fatalf("save: %v", err)
	}

	dirInfo, err := os.Stat(base)
	if err != nil {
		t.Fatalf("stat dir: %v", err)
	}
	if dirInfo.Mode().Perm() != 0o700 {
		t.Fatalf("want dir mode 0700, got %o", dirInfo.Mode().Perm())
	}

	fileInfo, err := os.Stat(filepath.Join(base, "config.json"))
	if err != nil {
		t.Fatalf("stat file: %v", err)
	}
	if fileInfo.Mode().Perm() != 0o600 {
		t.Fatalf("want file mode 0600, got %o", fileInfo.Mode().Perm())
	}
}
