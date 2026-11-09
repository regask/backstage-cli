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
	var posted []string
	cl := newPromoteFakeClient(&posted)
	msg := promoteCmd(cl, "staging", []string{"component:default/svc"})()
	res, ok := msg.(actionResultMsg)
	if !ok {
		t.Fatalf("want actionResultMsg, got %#v", msg)
	}
	if !res.OK {
		t.Fatalf("expected OK result, got %+v", res)
	}
	want := `{"templateRef":"template:default/promote-code","values":{"environment":"staging","services":["regask/svc"]}}`
	if len(posted) != 1 || posted[0] != want {
		t.Fatalf("posted = %v, want %s", posted, want)
	}
}

func TestPromoteCmdRejectsInvalidTarget(t *testing.T) {
	var posted []string
	msg := promoteCmd(newPromoteFakeClient(&posted), "development", []string{"component:default/svc"})()
	if _, ok := msg.(errMsg); !ok {
		t.Fatalf("want errMsg, got %#v", msg)
	}
	if len(posted) != 0 {
		t.Fatalf("launched a task for an invalid target: %v", posted)
	}
}

func TestReleaseCmdEmitsResult(t *testing.T) {
	var posted []string
	cl := newPromoteFakeClient(&posted)
	msg := releaseCmd(cl, "production")()
	res, ok := msg.(actionResultMsg)
	if !ok {
		t.Fatalf("want actionResultMsg, got %T", msg)
	}
	if !res.OK {
		t.Fatalf("expected OK result, got %+v", res)
	}
	want := `{"templateRef":"template:default/release","values":{"environment":"production"}}`
	if len(posted) != 1 || posted[0] != want {
		t.Fatalf("posted = %v, want %s", posted, want)
	}
}
