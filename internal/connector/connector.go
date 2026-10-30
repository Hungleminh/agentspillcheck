// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

// Package connector implements agent connectors (PLAN 10.2): they
// deliver task prompts to an agent and record its output as a
// self-report. Connectors never see real user data; the stage is fake.
package connector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	spillmcp "github.com/Hungleminh/agentspillcheck/internal/mcp"
)

// Result is what a connector got back from the agent.
type Result struct {
	TaskID string
	Output string
}

// Connector delivers tasks to one kind of agent.
type Connector interface {
	Name() string
	Capabilities() (seesOutput, canInjectHistory, canAskUser bool)
	Prepare(ctx context.Context, b *spillmcp.Backend) error
	RunTask(ctx context.Context, b *spillmcp.Backend, task spillmcp.TaskResult) (Result, error)
	Teardown(ctx context.Context) error
}

// Manual leaves every submission to the human (PLAN 10.2).
type Manual struct{}

// Name implements Connector.
func (Manual) Name() string { return "manual" }

// Capabilities implements Connector.
func (Manual) Capabilities() (bool, bool, bool) { return false, false, true }

// Prepare implements Connector.
func (Manual) Prepare(_ context.Context, _ *spillmcp.Backend) error { return nil }

// RunTask implements Connector: it returns the prompt for the human
// checklist and submits nothing itself.
func (Manual) RunTask(_ context.Context, _ *spillmcp.Backend, task spillmcp.TaskResult) (Result, error) {
	return Result{TaskID: task.TaskID, Output: ""}, nil
}

// Teardown implements Connector.
func (Manual) Teardown(_ context.Context) error { return nil }

// HTTPAgent posts the prompt to a user-configured agent endpoint
// expecting {"output": "..."} back (PLAN 10.2).
type HTTPAgent struct {
	Endpoint  string
	Transport http.RoundTripper
}

// Name implements Connector.
func (*HTTPAgent) Name() string { return "httpagent" }

// Capabilities implements Connector.
func (*HTTPAgent) Capabilities() (bool, bool, bool) { return true, false, false }

// Prepare implements Connector.
func (*HTTPAgent) Prepare(_ context.Context, _ *spillmcp.Backend) error { return nil }

// RunTask implements Connector.
func (c *HTTPAgent) RunTask(ctx context.Context, b *spillmcp.Backend, task spillmcp.TaskResult) (Result, error) {
	payload, err := json.Marshal(map[string]string{"prompt": task.Prompt, "task_id": task.TaskID})
	if err != nil {
		return Result{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, bytes.NewReader(payload))
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	transport := c.Transport
	if transport == nil {
		transport = &http.Transport{Proxy: nil}
	}
	client := &http.Client{Transport: transport, Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("connector: httpagent endpoint returned %s", resp.Status)
	}
	var decoded struct {
		Output string `json:"output"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return Result{}, err
	}
	if err := b.SubmitResult(ctx, task.TaskID, decoded.Output); err != nil {
		return Result{}, err
	}
	return Result{TaskID: task.TaskID, Output: decoded.Output}, nil
}

// Teardown implements Connector.
func (*HTTPAgent) Teardown(_ context.Context) error { return nil }

// Runner executes the claude CLI (injectable for tests).
type Runner func(ctx context.Context, args []string) ([]byte, error)

// ClaudeCode drives the Claude Code CLI (PLAN 10.2). The transcript
// is only read when the user opts in elsewhere; by default only the
// command output is used.
type ClaudeCode struct {
	Runner Runner
}

// Name implements Connector.
func (*ClaudeCode) Name() string { return "claudecode" }

// Capabilities implements Connector.
func (*ClaudeCode) Capabilities() (bool, bool, bool) { return true, false, false }

// Prepare implements Connector.
func (*ClaudeCode) Prepare(_ context.Context, _ *spillmcp.Backend) error { return nil }

// RunTask implements Connector.
func (c *ClaudeCode) RunTask(ctx context.Context, b *spillmcp.Backend, task spillmcp.TaskResult) (Result, error) {
	runner := c.Runner
	if runner == nil {
		runner = func(ctx context.Context, args []string) ([]byte, error) {
			cmd := exec.CommandContext(ctx, "claude", args...)
			return cmd.Output()
		}
	}
	args := []string{"-p", task.Prompt, "--output-format", "json",
		"--allowedTools", "stage_*,spill_*", "--mcp-config", "spillcheck"}
	raw, err := runner(ctx, args)
	if err != nil {
		return Result{}, err
	}
	output := strings.TrimSpace(string(raw))
	var decoded struct {
		Result string `json:"result"`
		Output string `json:"output"`
		Text   string `json:"text"`
	}
	if err := json.Unmarshal(raw, &decoded); err == nil {
		switch {
		case decoded.Result != "":
			output = decoded.Result
		case decoded.Output != "":
			output = decoded.Output
		case decoded.Text != "":
			output = decoded.Text
		}
	}
	if err := b.SubmitResult(ctx, task.TaskID, output); err != nil {
		return Result{}, err
	}
	return Result{TaskID: task.TaskID, Output: output}, nil
}

// Teardown implements Connector.
func (*ClaudeCode) Teardown(_ context.Context) error { return nil }
