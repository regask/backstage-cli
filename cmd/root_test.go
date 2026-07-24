package cmd

import (
	"bytes"
	"testing"
)

func TestRootShowsHelp(t *testing.T) {
	buf := &bytes.Buffer{}
	RootCmd.SetOut(buf)
	RootCmd.SetArgs([]string{"--help"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("backstage-regask")) {
		t.Fatalf("help missing binary name, got: %s", buf.String())
	}
}
