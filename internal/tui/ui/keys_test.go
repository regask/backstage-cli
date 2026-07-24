package ui

import "testing"

func TestDefaultKeysBound(t *testing.T) {
	k := DefaultKeys()
	if !k.Command.Enabled() || !k.Filter.Enabled() || !k.Quit.Enabled() {
		t.Fatal("core keys must be enabled")
	}
	if len(k.ShortHelp()) == 0 {
		t.Fatal("ShortHelp must list bindings")
	}
	if len(k.FullHelp()) == 0 {
		t.Fatal("FullHelp must list binding groups")
	}
}
