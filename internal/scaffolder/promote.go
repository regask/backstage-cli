package scaffolder

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

const PromoteTemplate = "template:default/promote-code"

// PromoteTargets mirrors PROMOTE_TARGETS in the portal
// (packages/app/src/modules/templates/gitopsServices.ts): the targets every
// service's promote.yml accepts. `development` is not one.
var PromoteTargets = []string{"staging", "pre-prod", "production"}

type SlugResolver interface {
	ProjectSlug(ctx context.Context, entityRef string) (string, error)
}

// PromoteValues builds the promote-code template parameters. Its
// regask:github:promote action dispatches each repo's promote.yml, so services
// are "owner/repo" slugs, not entity refs — deduped, since several components
// can build from one repo.
func PromoteValues(ctx context.Context, r SlugResolver, env string, serviceRefs []string) (map[string]any, error) {
	if !slices.Contains(PromoteTargets, env) {
		return nil, fmt.Errorf("invalid promote target %q (want one of: %s)", env, strings.Join(PromoteTargets, ", "))
	}
	slugs := make([]string, 0, len(serviceRefs))
	for _, ref := range serviceRefs {
		slug, err := r.ProjectSlug(ctx, ref)
		if err != nil {
			return nil, err
		}
		if !slices.Contains(slugs, slug) {
			slugs = append(slugs, slug)
		}
	}
	return map[string]any{"environment": env, "services": slugs}, nil
}
