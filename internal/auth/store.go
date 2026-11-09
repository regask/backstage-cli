package auth

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

type Config struct {
	PortalURL    string    `json:"portalUrl"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refreshToken,omitempty"`
	Expiry       time.Time `json:"expiry,omitempty"`
}

type Store struct{ dir string }

func NewStore(dir string) *Store { return &Store{dir: dir} }

// DefaultDir is ~/.config/backstage-regask (explicit, not os.UserConfigDir,
// which is ~/Library/Application Support on macOS).
func DefaultDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "backstage-regask"), nil
}

func (s *Store) path() string { return filepath.Join(s.dir, "config.json") }

func (s *Store) Load() (*Config, error) {
	b, err := os.ReadFile(s.path())
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) Save(c *Config) error {
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return err
	}
	// MkdirAll doesn't tighten perms on pre-existing dirs; enforce 0700 unconditionally.
	if err := os.Chmod(s.dir, 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	// Write-then-rename so a concurrent bsr process never reads a half-written
	// file mid token rotation. CreateTemp makes the file 0600.
	f, err := os.CreateTemp(s.dir, "config-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), s.path())
}

// Lock takes an exclusive cross-process lock serializing token refreshes, so
// two bsr processes never spend the same single-use refresh token.
func (s *Store) Lock() (func(), error) {
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(s.dir, "config.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, err
	}
	return func() { f.Close() }, nil
}

func (s *Store) Clear() error {
	err := os.Remove(s.path())
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
