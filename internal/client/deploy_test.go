package client

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestMatrixSendsServiceQuery(t *testing.T) {
	var gotURL string
	hc := &http.Client{Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
		gotURL = r.URL.String()
		return resp(200, `{"rows":[{"service":"svc","environments":{"prod":"v1"}}]}`), nil
	})}
	c := New("https://p", func() (string, error) { return "t", nil }, hc)
	m, err := c.Matrix(context.Background(), "svc", false)
	if err != nil {
		t.Fatalf("matrix: %v", err)
	}
	if !strings.Contains(gotURL, "service=svc") {
		t.Fatalf("url = %q", gotURL)
	}
	if len(m.Rows) != 1 || m.Rows[0].Environments["prod"] != "v1" {
		t.Fatalf("rows = %+v", m.Rows)
	}
}

func TestTicketLookupPostsBody(t *testing.T) {
	var gotBody string
	hc := &http.Client{Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		return resp(200, `{"results":{"ABC-1":["prod"]}}`), nil
	})}
	c := New("https://p", func() (string, error) { return "t", nil }, hc)
	out, err := c.TicketLookup(context.Background(), []string{"ABC-1"}, false)
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if !strings.Contains(gotBody, "ABC-1") {
		t.Fatalf("body = %q", gotBody)
	}
	if out.Results["ABC-1"][0] != "prod" {
		t.Fatalf("results = %+v", out.Results)
	}
}
