package auth

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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
	if err := os.WriteFile(s.path(), b, 0o600); err != nil {
		return err
	}
	// WriteFile doesn't tighten perms on pre-existing files; enforce 0600 unconditionally.
	return os.Chmod(s.path(), 0o600)
}

func (s *Store) Clear() error {
	err := os.Remove(s.path())
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
