package client

import (
	"context"
	"net/url"

	"github.com/regask/backstage-cli/internal/contracts"
)

func (c *Client) Matrix(ctx context.Context, service string, fresh bool) ([]contracts.MatrixRow, error) {
	q := url.Values{}
	if service != "" {
		q.Set("service", service)
	}
	var out []contracts.MatrixRow
	err := c.GetJSON(ctx, "/deploy-management/matrix", q, fresh, &out)
	return out, err
}

func (c *Client) Overlays(ctx context.Context, service string, fresh bool) (contracts.OverlayBundle, error) {
	q := url.Values{}
	q.Set("service", service)
	var out contracts.OverlayBundle
	err := c.GetJSON(ctx, "/deploy-management/service/overlays", q, fresh, &out)
	return out, err
}

func (c *Client) TicketLookup(ctx context.Context, tickets []string, fresh bool) (contracts.TicketLookupResult, error) {
	body := map[string]any{"tickets": tickets}
	var out contracts.TicketLookupResult
	path := "/deploy-management/ticket-lookup"
	if fresh {
		path += "?refresh=1"
	}
	err := c.PostJSON(ctx, path, body, &out)
	return out, err
}
