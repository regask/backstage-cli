package scaffolder

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

// Builders for the portal's scaffolder templates (templates/*/template.yaml in
// the backstage repo). Their actions dispatch per-repo workflows, so services
// are "owner/repo" slugs, not entity refs.
const (
	PromoteTemplate    = "template:default/promote-code"
	ReleaseTemplate    = "template:default/release"
	CherryPickTemplate = "template:default/cherry-pick"
)

// PromoteTargets mirrors PROMOTE_TARGETS in the portal
// (packages/app/src/modules/templates/gitopsServices.ts): the targets every
// service's promote.yml accepts. `development` is not one.
var PromoteTargets = []string{"staging", "pre-prod", "production"}

// ReleaseTargets is the release template's environment enum.
var ReleaseTargets = []string{"pre-prod", "production"}

// CherryPickBranches are the release branches the cherry-pick template targets;
// its environment enum is the suffix (preprod | prod).
var CherryPickBranches = []string{"release/preprod", "release/prod"}

type SlugResolver interface {
	ProjectSlug(ctx context.Context, entityRef string) (string, error)
}

// PromoteValues builds promote-code parameters. A blank tag lets the action
// auto-pick the latest.
func PromoteValues(ctx context.Context, r SlugResolver, env string, serviceRefs []string, tag string) (map[string]any, error) {
	if err := oneOf("promote target", env, PromoteTargets); err != nil {
		return nil, err
	}
	slugs, err := slugsFor(ctx, r, serviceRefs)
	if err != nil {
		return nil, err
	}
	values := map[string]any{"environment": env, "services": slugs}
	if tag != "" {
		values["tag"] = tag
	}
	return values, nil
}

// ReleaseValues builds release parameters. sharedActionsRelease 0 leaves the
// shared-actions version unchanged; the template only applies it to production.
func ReleaseValues(ctx context.Context, r SlugResolver, env string, excludeRefs []string, sharedActionsRelease int) (map[string]any, error) {
	if err := oneOf("release environment", env, ReleaseTargets); err != nil {
		return nil, err
	}
	values := map[string]any{"environment": env}
	if len(excludeRefs) > 0 {
		slugs, err := slugsFor(ctx, r, excludeRefs)
		if err != nil {
			return nil, err
		}
		values["excludeServices"] = slugs
	}
	if sharedActionsRelease != 0 {
		if env != "production" || sharedActionsRelease < 1 {
			return nil, fmt.Errorf("shared-actions release version must be a positive number and applies to production only")
		}
		values["sharedActionsRelease"] = sharedActionsRelease
	}
	return values, nil
}

// CherryPickValues builds cherry-pick parameters from ticket refs (1234 or
// REG-1234) and a target release branch.
func CherryPickValues(tickets []string, branch string) (map[string]any, error) {
	if len(tickets) == 0 {
		return nil, fmt.Errorf("at least one ticket is required (e.g. REG-12345)")
	}
	if err := oneOf("cherry-pick branch", branch, CherryPickBranches); err != nil {
		return nil, err
	}
	return map[string]any{"tickets": tickets, "environment": strings.TrimPrefix(branch, "release/")}, nil
}

func oneOf(what, v string, allowed []string) error {
	if !slices.Contains(allowed, v) {
		return fmt.Errorf("invalid %s %q (want one of: %s)", what, v, strings.Join(allowed, ", "))
	}
	return nil
}

// slugsFor resolves refs to slugs, deduped: several components can build from
// one repo.
func slugsFor(ctx context.Context, r SlugResolver, refs []string) ([]string, error) {
	slugs := make([]string, 0, len(refs))
	for _, ref := range refs {
		slug, err := r.ProjectSlug(ctx, ref)
		if err != nil {
			return nil, err
		}
		if !slices.Contains(slugs, slug) {
			slugs = append(slugs, slug)
		}
	}
	return slugs, nil
}
