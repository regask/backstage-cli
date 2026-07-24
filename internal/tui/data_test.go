package tui

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/regask/backstage-cli/internal/client"
)

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func newFakeClient(body string) *client.Client {
	hc := &http.Client{Transport: rtFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{}}, nil
	})}
	return client.New("https://p", func() (string, error) { return "t", nil }, hc)
}

func TestLoadServicesEmitsRows(t *testing.T) {
	cl := newFakeClient(`[{"serviceRef":"component:default/svc","serviceName":"svc","envs":{"production":{"tag":"v1","syncStatus":"Synced","healthStatus":"Healthy"}}}]`)
	msg := loadServices(cl, false)()
	loaded, ok := msg.(servicesLoadedMsg)
	if !ok {
		t.Fatalf("want servicesLoadedMsg, got %T", msg)
	}
	if len(loaded.Rows) != 1 || loaded.Rows[0].Envs["production"].Tag != "v1" {
		t.Fatalf("rows = %+v", loaded.Rows)
	}
}
