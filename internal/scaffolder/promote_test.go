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
		[]string{"component:default/web-frontend", "component:default/storybook", "component:default/alert"})
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
		if _, err := PromoteValues(context.Background(), slugMap{}, env, nil); err == nil {
			t.Fatalf("env %q: expected error", env)
		}
	}
}

func TestPromoteValuesPropagatesLookupError(t *testing.T) {
	if _, err := PromoteValues(context.Background(), slugMap{}, "staging", []string{"component:default/x"}); err == nil {
		t.Fatalf("expected lookup error")
	}
}
