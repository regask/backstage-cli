package client

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestParseApprovalID(t *testing.T) {
	cases := map[string]string{
		"abc-123": "abc-123",
		"https://backstage.regask.com/approvals/abc-123":       "abc-123",
		"https://backstage.regask.com/approvals/abc-123?x=1#y": "abc-123",
	}
	for in, want := range cases {
		got, err := ParseApprovalID(in)
		if err != nil || got != want {
			t.Fatalf("ParseApprovalID(%q) = %q, %v", in, got, err)
		}
	}
}

func TestDecideApprovalPostsToApprovePath(t *testing.T) {
	var gotPath string
	hc := &http.Client{Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
		gotPath = r.URL.Path
		return resp(200, `{}`), nil
	})}
	c := New("https://p", func() (string, error) { return "t", nil }, hc)
	if err := c.DecideApproval(context.Background(), "abc", true); err != nil {
		t.Fatalf("decide: %v", err)
	}
	if !strings.HasSuffix(gotPath, "/approvals/requests/abc/approve") {
		t.Fatalf("path = %q", gotPath)
	}
}

func TestGetApprovalReadsResultURLAndTaskID(t *testing.T) {
	hc := &http.Client{Transport: rtFunc(func(*http.Request) (*http.Response, error) {
		return resp(200, `{"id":"abc","kind":"release-publish","status":"pending","title":"Publish svc","resultUrl":"https://github.com/regask/svc/releases/tag/v1","payload":{"taskId":"task-9"}}`), nil
	})}
	c := New("https://p", func() (string, error) { return "t", nil }, hc)
	r, err := c.GetApproval(context.Background(), "abc")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if r.ResultURL != "https://github.com/regask/svc/releases/tag/v1" {
		t.Fatalf("resultUrl = %q", r.ResultURL)
	}
	if r.TaskID() != "task-9" {
		t.Fatalf("taskID = %q", r.TaskID())
	}
}
