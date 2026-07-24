package main

import (
	"fmt"
	"os"

	"github.com/regask/backstage-cli/cmd"
)

// version is overridden at build time via -X main.version={{.Version}} (see .goreleaser.yaml).
var version = "dev"

func main() {
	cmd.RootCmd.Version = version
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
