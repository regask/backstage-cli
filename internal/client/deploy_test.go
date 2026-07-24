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
		return resp(200, `[{"serviceRef":"svc","serviceName":"svc","envs":{"production":{"tag":"v1"}}}]`), nil
	})}
	c := New("https://p", func() (string, error) { return "t", nil }, hc)
	m, err := c.Matrix(context.Background(), "svc", false)
	if err != nil {
		t.Fatalf("matrix: %v", err)
	}
	if !strings.Contains(gotURL, "service=svc") {
		t.Fatalf("url = %q", gotURL)
	}
	if len(m) != 1 || m[0].Envs["production"].Tag != "v1" {
		t.Fatalf("rows = %+v", m)
	}
}

func TestTicketLookupPostsBody(t *testing.T) {
	var gotBody string
	hc := &http.Client{Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		return resp(200, `{"services":[{"serviceRef":"component:default/svc","slug":"svc","count":1,"deployedEnvs":["production"],"commits":[]}],"notFound":[],"searchRateLimited":false}`), nil
	})}
	c := New("https://p", func() (string, error) { return "t", nil }, hc)
	out, err := c.TicketLookup(context.Background(), []string{"ABC-1"}, false)
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if !strings.Contains(gotBody, "ABC-1") {
		t.Fatalf("body = %q", gotBody)
	}
	if out.Services[0].DeployedEnvs[0] != "production" {
		t.Fatalf("services = %+v", out.Services)
	}
}
