package scaffolder

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strconv"
)

// TODO(execution): confirm these endpoints against the running backend —
// the Backstage default scaffolder API is /scaffolder/v2/tasks (POST) and
// /scaffolder/v2/tasks/:id/events?after=N (GET, long-poll).
type Doer interface {
	PostJSON(ctx context.Context, path string, body, out any) error
	GetJSON(ctx context.Context, path string, q url.Values, fresh bool, out any) error
}

type event struct {
	ID     int    `json:"id"`
	Type   string `json:"type"`
	Status string `json:"status"`
	Body   struct {
		Message string `json:"message"`
	} `json:"body"`
}

// Launch creates a scaffolder task and returns its id.
func Launch(ctx context.Context, d Doer, templateRef string, values map[string]any) (string, error) {
	body := map[string]any{"templateRef": templateRef, "values": values}
	var out struct {
		ID string `json:"id"`
	}
	if err := d.PostJSON(ctx, "/scaffolder/v2/tasks", body, &out); err != nil {
		return "", err
	}
	if out.ID == "" {
		return "", fmt.Errorf("scaffolder did not return a task id")
	}
	return out.ID, nil
}

// Stream long-polls the task's events, writing log lines to out, until a
// completion event arrives. Returns the terminal status.
func Stream(ctx context.Context, d Doer, taskID string, out io.Writer) (string, error) {
	after := 0
	for {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		q := url.Values{}
		q.Set("after", strconv.Itoa(after))
		var batch []event
		if err := d.GetJSON(ctx, "/scaffolder/v2/tasks/"+url.PathEscape(taskID)+"/events", q, false, &batch); err != nil {
			return "", err
		}
		for _, e := range batch {
			if e.ID > after {
				after = e.ID
			}
			if e.Body.Message != "" {
				fmt.Fprintln(out, e.Body.Message)
			}
			if e.Type == "completion" {
				status := e.Status
				if status == "" {
					status = "completed"
				}
				return status, nil
			}
		}
	}
}
