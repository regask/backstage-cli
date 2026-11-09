package scaffolder

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

type slugMap map[string]string

func (m slugMap) ProjectSlug(_ context.Context, ref string) (string, error) {
	if s, ok := m[ref]; ok {
		return s, nil
	}
	return "", fmt.Errorf("unknown %s", ref)
}

func TestPromoteValuesMatchesTemplateParams(t *testing.T) {
	r := slugMap{
		"component:default/web-frontend": "regask/web-frontend",
		"component:default/storybook":    "regask/web-frontend",
		"component:default/alert":        "regask/alert",
	}
	got, err := PromoteValues(context.Background(), r, "staging",
		[]string{"component:default/web-frontend", "component:default/storybook", "component:default/alert"}, "")
	if err != nil {
		t.Fatalf("values: %v", err)
	}
	want := map[string]any{"environment": "staging", "services": []string{"regask/web-frontend", "regask/alert"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestPromoteValuesRejectsInvalidTarget(t *testing.T) {
	for _, env := range []string{"development", "prod", ""} {
		if _, err := PromoteValues(context.Background(), slugMap{}, env, nil, ""); err == nil {
			t.Fatalf("env %q: expected error", env)
		}
	}
}

func TestPromoteValuesPropagatesLookupError(t *testing.T) {
	if _, err := PromoteValues(context.Background(), slugMap{}, "staging", []string{"component:default/x"}, ""); err == nil {
		t.Fatalf("expected lookup error")
	}
}

func TestPromoteValuesSendsTagOnlyWhenSet(t *testing.T) {
	r := slugMap{"component:default/a": "regask/a"}
	got, _ := PromoteValues(context.Background(), r, "staging", []string{"component:default/a"}, "v1.2.3")
	if got["tag"] != "v1.2.3" {
		t.Fatalf("tag = %v", got["tag"])
	}
	got, _ = PromoteValues(context.Background(), r, "staging", []string{"component:default/a"}, "")
	if _, ok := got["tag"]; ok {
		t.Fatalf("blank tag sent: %v", got)
	}
}

func TestReleaseValuesMatchesTemplateParams(t *testing.T) {
	r := slugMap{"component:default/a": "regask/a"}
	got, err := ReleaseValues(context.Background(), r, "production", []string{"component:default/a"}, 42)
	if err != nil {
		t.Fatalf("values: %v", err)
	}
	want := map[string]any{"environment": "production", "excludeServices": []string{"regask/a"}, "sharedActionsRelease": 42}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	got, _ = ReleaseValues(context.Background(), r, "pre-prod", nil, 0)
	if !reflect.DeepEqual(got, map[string]any{"environment": "pre-prod"}) {
		t.Fatalf("got %v", got)
	}
}

func TestReleaseValuesRejectsInvalidInput(t *testing.T) {
	cases := []struct {
		env     string
		version int
	}{{"staging", 0}, {"prod", 0}, {"pre-prod", 3}, {"production", -1}}
	for _, c := range cases {
		if _, err := ReleaseValues(context.Background(), slugMap{}, c.env, nil, c.version); err == nil {
			t.Fatalf("%+v: expected error", c)
		}
	}
}

func TestCherryPickValuesMapsBranchToEnvironment(t *testing.T) {
	got, err := CherryPickValues([]string{"REG-1", "2"}, "release/prod")
	if err != nil {
		t.Fatalf("values: %v", err)
	}
	want := map[string]any{"tickets": []string{"REG-1", "2"}, "environment": "prod"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if _, err := CherryPickValues(nil, "release/prod"); err == nil {
		t.Fatalf("no tickets: expected error")
	}
	if _, err := CherryPickValues([]string{"REG-1"}, "main"); err == nil {
		t.Fatalf("bad branch: expected error")
	}
}
