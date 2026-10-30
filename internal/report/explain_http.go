// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenAIClient explains reports through an OpenAI-compatible chat
// endpoint with the user's own key (ADR-0005). The key is held only
// in memory and never logged. Default off: callers must opt in.
type OpenAIClient struct {
	BaseURL   string // e.g. https://api.openai.com/v1
	APIKey    string
	Model     string
	Transport http.RoundTripper
}

// Complete sends one chat completion and returns the text.
func (c *OpenAIClient) Complete(ctx context.Context, prompt string) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("report: explain requires the user's API key")
	}
	base := c.BaseURL
	if base == "" {
		base = "https://api.openai.com/v1"
	}
	model := c.Model
	if model == "" {
		model = "gpt-4o-mini"
	}
	body, err := json.Marshal(map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	transport := c.Transport
	if transport == nil {
		transport = &http.Transport{Proxy: nil}
	}
	client := &http.Client{Transport: transport, Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("report: explain provider returned %s", resp.Status)
	}
	var decoded struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return "", err
	}
	if len(decoded.Choices) == 0 {
		return "", fmt.Errorf("report: explain provider returned no choices")
	}
	return decoded.Choices[0].Message.Content, nil
}
