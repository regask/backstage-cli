package az

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type Runner interface {
	GetSecret(ctx context.Context, vault, name string) (string, error)
}

type CLI struct {
	Exec func(ctx context.Context, name string, args ...string) ([]byte, error)
}

func NewCLI() *CLI {
	return &CLI{Exec: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return exec.CommandContext(ctx, name, args...).Output()
	}}
}

func (c *CLI) GetSecret(ctx context.Context, vault, name string) (string, error) {
	out, err := c.Exec(ctx, "az", "keyvault", "secret", "show",
		"--vault-name", vault, "--name", name, "--query", "value", "-o", "tsv")
	if err != nil {
		return "", fmt.Errorf("az keyvault read failed for %q in %q (is `az login` active?): %w", name, vault, err)
	}
	return strings.TrimSpace(string(out)), nil
}
