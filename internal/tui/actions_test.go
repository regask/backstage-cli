package tui

import (
	"testing"
)

func TestApproveCmdEmitsResult(t *testing.T) {
	cl := newFakeClient(`{}`) // POST approve → 200 {}
	msg := approveCmd(cl, "abc", true)()
	res, ok := msg.(actionResultMsg)
	if !ok {
		t.Fatalf("want actionResultMsg, got %T", msg)
	}
	if !res.OK {
		t.Fatalf("expected OK result, got %+v", res)
	}
}

func TestApproveCmdRejectEmitsResult(t *testing.T) {
	cl := newFakeClient(`{}`)
	msg := approveCmd(cl, "xyz", false)()
	res, ok := msg.(actionResultMsg)
	if !ok {
		t.Fatalf("want actionResultMsg, got %T", msg)
	}
	if !res.OK || res.Text == "" {
		t.Fatalf("expected OK result with text, got %+v", res)
	}
}

func TestPromoteCmdEmitsResult(t *testing.T) {
	cl := newFakeClient(`{"id":"task-1"}`)
	msg := promoteCmd(cl, "staging", []string{"component:default/svc"})()
	res, ok := msg.(actionResultMsg)
	if !ok {
		t.Fatalf("want actionResultMsg, got %T", msg)
	}
	if !res.OK {
		t.Fatalf("expected OK result, got %+v", res)
	}
}

func TestReleaseCmdEmitsResult(t *testing.T) {
	cl := newFakeClient(`{"id":"task-2"}`)
	msg := releaseCmd(cl, "production")()
	res, ok := msg.(actionResultMsg)
	if !ok {
		t.Fatalf("want actionResultMsg, got %T", msg)
	}
	if !res.OK {
		t.Fatalf("expected OK result, got %+v", res)
	}
}
