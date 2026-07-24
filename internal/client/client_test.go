package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func resp(code int, body string) *http.Response {
	return &http.Response{StatusCode: code, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{}}
}

func TestGetJSONInjectsTokenAndFresh(t *testing.T) {
	var gotAuth, gotURL string
	hc := &http.Client{Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
		gotAuth = r.Header.Get("Authorization")
		gotURL = r.URL.String()
		return resp(200, `{"ok":true}`), nil
	})}
	c := New("https://portal.example/", func() (string, error) { return "tok123", nil }, hc)

	var out struct {
		OK bool `json:"ok"`
	}
	if err := c.GetJSON(context.Background(), "/deploy-management/matrix", nil, true, &out); err != nil {
		t.Fatalf("get: %v", err)
	}
	if gotAuth != "Bearer tok123" {
		t.Fatalf("auth header = %q", gotAuth)
	}
	if !strings.Contains(gotURL, "/api/deploy-management/matrix") || !strings.Contains(gotURL, "refresh=1") {
		t.Fatalf("url = %q", gotURL)
	}
	if !out.OK {
		t.Fatalf("decode failed: %+v", out)
	}
}

func TestUnauthorizedMapsToSentinel(t *testing.T) {
	hc := &http.Client{Transport: rtFunc(func(*http.Request) (*http.Response, error) {
		return resp(401, `unauthorized`), nil
	})}
	c := New("https://portal.example", func() (string, error) { return "t", nil }, hc)
	err := c.GetJSON(context.Background(), "/x", nil, false, nil)
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("want ErrUnauthorized, got %v", err)
	}
}
