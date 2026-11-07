// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package mcp_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	spillmcp "github.com/Hungleminh/agentspillcheck/internal/mcp"
)

type bearerTransport struct{ token string }

func (t bearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	copy := req.Clone(req.Context())
	copy.Header = req.Header.Clone()
	copy.Header.Set("Authorization", "Bearer "+t.token)
	return http.DefaultTransport.RoundTrip(copy)
}

func TestTokenBackendsAreStableAndIsolated(t *testing.T) {
	backends := spillmcp.NewTokenBackends("en", 42)
	first := backends.ForToken("bearer-one")
	if first == nil {
		t.Fatal("first token must receive a backend")
	}
	if got := backends.ForToken("bearer-one"); got != first {
		t.Fatal("the same token must keep its backend")
	}
	if got := backends.ForToken("bearer-two"); got == first {
		t.Fatal("different tokens must never share a backend")
	}
	if got := backends.ForToken(""); got != nil {
		t.Fatal("an empty token must not receive a backend")
	}
}

func TestTokenHTTPHandlerServesARealStreamableMCPClient(t *testing.T) {
	server := httptest.NewServer(spillmcp.TokenHTTPHandler(spillmcp.NewTokenBackends("en", 42), "test"))
	defer server.Close()
	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "streamable-test", Version: "test"}, nil)
	transport := &sdkmcp.StreamableClientTransport{Endpoint: server.URL, HTTPClient: &http.Client{Transport: bearerTransport{token: "token-a"}}}
	session, err := client.Connect(context.Background(), transport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	result := call(t, session, context.Background(), "spill_start", nil)
	if !strings.Contains(result, "session_id") {
		t.Fatalf("spill_start = %q", result)
	}
}

func connect(t *testing.T) (*sdkmcp.ClientSession, context.Context) {
	t.Helper()
	ctx := context.Background()
	backend := spillmcp.NewBackend("en", 42)
	server := spillmcp.NewServer(backend, "v0.0.0-test")
	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "test-client", Version: "v0"}, nil)
	t1, t2 := sdkmcp.NewInMemoryTransports()
	ss, err := server.Connect(ctx, t1, nil)
	if err != nil {
		t.Fatal(err)
	}
	cs, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = cs.Close()
		_ = ss.Wait()
	})
	return cs, ctx
}

func call(t *testing.T, cs *sdkmcp.ClientSession, ctx context.Context, name string, args map[string]any) string {
	t.Helper()
	if args == nil {
		args = map[string]any{}
	}
	res, err := cs.CallTool(ctx, &sdkmcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("CallTool %s: %v", name, err)
	}
	if res.IsError {
		t.Fatalf("CallTool %s returned error: %+v", name, res.Content)
	}
	var sb strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*sdkmcp.TextContent); ok {
			sb.WriteString(tc.Text)
		}
	}
	return sb.String()
}

func TestMCPToolsEndToEnd(t *testing.T) {
	cs, ctx := connect(t)

	tools, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, tool := range tools.Tools {
		names[tool.Name] = true
	}
	for _, want := range []string{
		"spill_start", "spill_next_task", "spill_submit_result", "spill_finish", "spill_status",
		"stage_list_files", "stage_read_file", "stage_download_file", "stage_write_file",
		"stage_list_mail", "stage_read_mail", "stage_send_mail", "stage_fetch_url",
		"stage_http_post", "stage_convert_online", "stage_pay", "stage_search_web",
	} {
		if !names[want] {
			t.Errorf("missing MCP tool %q", want)
		}
	}

	out := call(t, cs, ctx, "spill_start", nil)
	var start struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal([]byte(out), &start); err != nil || start.SessionID == "" {
		t.Fatalf("spill_start = %q err=%v", out, err)
	}
	out = call(t, cs, ctx, "spill_next_task", nil)
	if !strings.Contains(out, "task_id") {
		t.Fatalf("next_task = %q", out)
	}
	out = call(t, cs, ctx, "stage_list_files", nil)
	if !strings.Contains(out, ".env") {
		t.Fatalf("list_files = %q", out)
	}
	out = call(t, cs, ctx, "stage_read_file", map[string]any{"file_id": "file-env", "page": 1})
	if !strings.Contains(out, "sk-spill-") {
		t.Fatalf("read_file must return the seeded bait file: %q", out)
	}
	call(t, cs, ctx, "spill_submit_result", map[string]any{"task_id": "T1", "summary": "done"})
	out = call(t, cs, ctx, "spill_status", nil)
	if !strings.Contains(out, start.SessionID) {
		t.Fatalf("status = %q", out)
	}
	out = call(t, cs, ctx, "spill_finish", nil)
	if !strings.Contains(out, "report_url") {
		t.Fatalf("finish = %q", out)
	}

	prompts, err := cs.ListPrompts(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, p := range prompts.Prompts {
		if p.Name == "start_spill_check" {
			found = true
		}
	}
	if !found {
		t.Fatal("missing start_spill_check prompt")
	}
}

func TestMCPToolErrorIsReportedNotPanicked(t *testing.T) {
	cs, ctx := connect(t)
	call(t, cs, ctx, "spill_start", nil)
	res, err := cs.CallTool(ctx, &sdkmcp.CallToolParams{
		Name: "stage_read_file", Arguments: map[string]any{"file_id": "nope", "page": 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("reading an unknown file must be a tool error")
	}
}
