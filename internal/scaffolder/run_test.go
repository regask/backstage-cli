package scaffolder

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"testing"
)

type fakeDoer struct {
	postOut  string
	events   [][]byte // successive event-batch responses
	eventIdx int
	task     string // GET /scaffolder/v2/tasks/:id response
}

func (f *fakeDoer) PostJSON(ctx context.Context, path string, body, out any) error {
	return json.Unmarshal([]byte(f.postOut), out)
}
func (f *fakeDoer) GetJSON(ctx context.Context, path string, q url.Values, fresh bool, out any) error {
	if !strings.HasSuffix(path, "/events") {
		return json.Unmarshal([]byte(f.task), out)
	}
	b := f.events[f.eventIdx]
	if f.eventIdx < len(f.events)-1 {
		f.eventIdx++
	}
	return json.Unmarshal(b, out)
}

func TestLaunchReturnsTaskID(t *testing.T) {
	d := &fakeDoer{postOut: `{"id":"task-1"}`}
	id, err := Launch(context.Background(), d, "template:default/promote-code", map[string]any{"service": "svc"})
	if err != nil || id != "task-1" {
		t.Fatalf("id=%q err=%v", id, err)
	}
}

func TestStreamPrintsLogAndReturnsStatus(t *testing.T) {
	d := &fakeDoer{events: [][]byte{
		[]byte(`[{"id":1,"type":"log","body":{"message":"starting"}}]`),
		[]byte(`[{"id":2,"type":"completion","body":{"message":"Run completed with status: completed"}}]`),
	}, task: `{"id":"task-1","status":"completed"}`}
	buf := &bytes.Buffer{}
	status, err := Stream(context.Background(), d, "task-1", buf)
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	if status != "completed" {
		t.Fatalf("status = %q", status)
	}
	if !bytes.Contains(buf.Bytes(), []byte("starting")) {
		t.Fatalf("log = %q", buf.String())
	}
}

// The completion event carries no status, so a failed run must be read from
// the task rather than defaulting to success.
func TestStreamReportsFailedTask(t *testing.T) {
	d := &fakeDoer{events: [][]byte{
		[]byte(`[{"id":1,"type":"completion","body":{"message":"Run completed with status: failed"}}]`),
	}, task: `{"id":"task-1","status":"failed"}`}
	status, err := Stream(context.Background(), d, "task-1", &bytes.Buffer{})
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	if status != "failed" {
		t.Fatalf("status = %q, want failed", status)
	}
}
