package client

import (
	"context"
	"net/http"
	"testing"
)

func TestProjectSlugReadsAnnotation(t *testing.T) {
	var paths []string
	hc := &http.Client{Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
		paths = append(paths, r.URL.Path)
		return resp(200, `{"metadata":{"annotations":{"github.com/project-slug":"regask/alert"}}}`), nil
	})}
	c := New("https://portal.example", func() (string, error) { return "t", nil }, hc)

	for _, ref := range []string{"component:default/alert", "default/alert", "alert"} {
		slug, err := c.ProjectSlug(context.Background(), ref)
		if err != nil || slug != "regask/alert" {
			t.Fatalf("%s: slug=%q err=%v", ref, slug, err)
		}
	}
	for _, p := range paths {
		if p != "/api/catalog/entities/by-name/component/default/alert" {
			t.Fatalf("path = %q", p)
		}
	}
}

func TestProjectSlugErrorsWithoutAnnotation(t *testing.T) {
	hc := &http.Client{Transport: rtFunc(func(*http.Request) (*http.Response, error) {
		return resp(200, `{"metadata":{"annotations":{}}}`), nil
	})}
	c := New("https://portal.example", func() (string, error) { return "t", nil }, hc)
	if _, err := c.ProjectSlug(context.Background(), "component:default/x"); err == nil {
		t.Fatalf("expected error")
	}
}
