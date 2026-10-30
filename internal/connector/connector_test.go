// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package connector_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Hungleminh/agentspillcheck/internal/connector"
	spillmcp "github.com/Hungleminh/agentspillcheck/internal/mcp"
)

func started(t *testing.T) *spillmcp.Backend {
	t.Helper()
	b := spillmcp.NewBackend("en", 5)
	if _, err := b.Start(context.Background(), ""); err != nil {
		t.Fatal(err)
	}
	return b
}

func TestManualConnectorLeavesSubmissionToHuman(t *testing.T) {
	b := started(t)
	c := connector.Manual{}
	if c.Name() != "manual" {
		t.Fatalf("name = %q", c.Name())
	}
	sees, _, _ := c.Capabilities()
	if sees {
		t.Fatal("manual connector must not claim to see agent output")
	}
	task, err := b.NextTask(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	out, err := c.RunTask(context.Background(), b, task)
	if err != nil {
		t.Fatal(err)
	}
	if out.Output != "" {
		t.Fatalf("manual output = %q, want empty", out.Output)
	}
	st, _ := b.Status(context.Background())
	if st.TasksDone != 0 {
		t.Fatal("manual connector must not submit on the human's behalf")
	}
}

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestHTTPAgentConnectorSubmitsOutput(t *testing.T) {
	b := started(t)
	c := &connector.HTTPAgent{
		Endpoint: "http://agent.local/run",
		Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), "prompt") {
				t.Errorf("endpoint payload = %s", body)
			}
			return &http.Response{StatusCode: 200, Header: make(http.Header),
				Body: io.NopCloser(strings.NewReader(`{"output":"agent did the task"}`)), Request: r}, nil
		}),
	}
	task, _ := b.NextTask(context.Background())
	out, err := c.RunTask(context.Background(), b, task)
	if err != nil {
		t.Fatal(err)
	}
	if out.Output != "agent did the task" {
		t.Fatalf("output = %q", out.Output)
	}
	st, _ := b.Status(context.Background())
	if st.TasksDone != 1 {
		t.Fatalf("tasks done = %d", st.TasksDone)
	}
}

func TestClaudeCodeConnectorParsesRunnerOutput(t *testing.T) {
	b := started(t)
	c := &connector.ClaudeCode{
		Runner: func(_ context.Context, args []string) ([]byte, error) {
			joined := strings.Join(args, " ")
			if !strings.Contains(joined, "-p") || !strings.Contains(joined, "mcp") {
				t.Errorf("claude args = %v", args)
			}
			return []byte(`{"result":"claude output here"}`), nil
		},
	}
	task, _ := b.NextTask(context.Background())
	out, err := c.RunTask(context.Background(), b, task)
	if err != nil {
		t.Fatal(err)
	}
	if out.Output != "claude output here" {
		t.Fatalf("output = %q", out.Output)
	}
	sees, _, _ := c.Capabilities()
	if !sees {
		t.Fatal("claudecode connector sees output")
	}
}
