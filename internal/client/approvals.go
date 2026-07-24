package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/regask/backstage-cli/internal/contracts"
)

// ParseApprovalID accepts a bare id or a portal link like
// https://host/approvals/<id>[?query][#frag] and returns the id.
func ParseApprovalID(linkOrID string) (string, error) {
	s := strings.TrimSpace(linkOrID)
	if !strings.Contains(s, "/") {
		return s, nil
	}
	u, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	for i, p := range parts {
		if p == "approvals" && i+1 < len(parts) {
			return parts[i+1], nil
		}
	}
	return "", fmt.Errorf("could not find approval id in %q", linkOrID)
}

func (c *Client) GetApproval(ctx context.Context, id string) (contracts.ApprovalRequest, error) {
	// GET /requests/:id wraps the request in a { "request": {...} } envelope.
	var out struct {
		Request contracts.ApprovalRequest `json:"request"`
	}
	err := c.GetJSON(ctx, "/approvals/requests/"+url.PathEscape(id), nil, false, &out)
	return out.Request, err
}

func (c *Client) DecideApproval(ctx context.Context, id string, approve bool) error {
	action := "reject"
	if approve {
		action = "approve"
	}
	return c.PostJSON(ctx, "/approvals/requests/"+url.PathEscape(id)+"/"+action, map[string]any{}, nil)
}
