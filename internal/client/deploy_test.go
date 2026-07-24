package client

import (
	"context"
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
