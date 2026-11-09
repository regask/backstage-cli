package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// ProjectSlug returns the entity's github.com/project-slug annotation
// ("owner/repo"). A ref without kind or namespace defaults to
// component / default.
func (c *Client) ProjectSlug(ctx context.Context, entityRef string) (string, error) {
	kind, rest := "component", entityRef
	if k, r, ok := strings.Cut(entityRef, ":"); ok {
		kind, rest = k, r
	}
	namespace, name := "default", rest
	if ns, n, ok := strings.Cut(rest, "/"); ok {
		namespace, name = ns, n
	}
	var e struct {
		Metadata struct {
			Annotations map[string]string `json:"annotations"`
		} `json:"metadata"`
	}
	path := "/catalog/entities/by-name/" + url.PathEscape(kind) + "/" + url.PathEscape(namespace) + "/" + url.PathEscape(name)
	if err := c.GetJSON(ctx, path, nil, false, &e); err != nil {
		return "", fmt.Errorf("look up %s: %w", entityRef, err)
	}
	slug := e.Metadata.Annotations["github.com/project-slug"]
	if slug == "" {
		return "", fmt.Errorf("%s has no github.com/project-slug annotation", entityRef)
	}
	return slug, nil
}
