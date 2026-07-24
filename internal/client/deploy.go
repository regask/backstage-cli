package client

import (
	"context"
	"net/url"

	"github.com/regask/backstage-cli/internal/contracts"
)

func (c *Client) Matrix(ctx context.Context, service string, fresh bool) (contracts.MatrixResponse, error) {
	q := url.Values{}
	if service != "" {
		q.Set("service", service)
	}
	var out contracts.MatrixResponse
	err := c.GetJSON(ctx, "/deploy-management/matrix", q, fresh, &out)
	return out, err
}

func (c *Client) Overlays(ctx context.Context, service string, fresh bool) (contracts.OverlayResponse, error) {
	q := url.Values{}
	q.Set("service", service)
	var out contracts.OverlayResponse
	err := c.GetJSON(ctx, "/deploy-management/service/overlays", q, fresh, &out)
	return out, err
}

func (c *Client) TicketLookup(ctx context.Context, tickets []string, fresh bool) (contracts.TicketLookupResponse, error) {
	body := map[string]any{"tickets": tickets}
	var out contracts.TicketLookupResponse
	path := "/deploy-management/ticket-lookup"
	if fresh {
		path += "?refresh=1"
	}
	err := c.PostJSON(ctx, path, body, &out)
	return out, err
}
