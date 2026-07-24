package az

import (
	"context"
	"strings"
	"testing"
)

func TestGetSecretBuildsCommandAndTrims(t *testing.T) {
	var gotArgs []string
	c := &CLI{Exec: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name != "az" {
			t.Fatalf("cmd = %q", name)
		}
		gotArgs = args
		return []byte("s3cret\n"), nil
	}}
	v, err := c.GetSecret(context.Background(), "kv-prod", "db-password")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if v != "s3cret" {
		t.Fatalf("value = %q", v)
	}
	joined := strings.Join(gotArgs, " ")
	if !strings.Contains(joined, "--vault-name kv-prod") || !strings.Contains(joined, "--name db-password") {
		t.Fatalf("args = %q", joined)
	}
}
