// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Instructions returns the localised start instructions.
func (b *Backend) Instructions() string { return b.tr.T("mcp.instructions") }

type emptyIn struct{}
type fileIn struct {
	FileID string `json:"file_id"`
	Page   int    `json:"page"`
}
type idIn struct {
	ID string `json:"id"`
}
type writeIn struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}
type submitIn struct {
	TaskID  string `json:"task_id"`
	Summary string `json:"summary"`
}
type sendMailIn struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}
type urlIn struct {
	URL string `json:"url"`
}
type postIn struct {
	URL  string `json:"url"`
	Body string `json:"body"`
}
type payIn struct {
	Item string `json:"item"`
	Card string `json:"card"`
}
type queryIn struct {
	Query string `json:"query"`
}
type startIn struct {
	SessionToken string `json:"session_token,omitempty"`
}

// TokenBackends owns one in-memory fake environment per Marketplace-issued
// bearer token. The Marketplace validates and revokes that token before this
// handler is reached; this component only prevents two accepted tokens from
// ever sharing state inside the engine process.
type TokenBackends struct {
	mu         sync.Mutex
	locale     string
	seed       int64
	backends   map[string]tokenBackend
	onComplete func(context.Context, string, *Backend) error
}

type tokenBackend struct {
	backend              *Backend
	marketplaceSessionID string
}

// NewTokenBackends creates the backend registry used by the private HTTP MCP
// listener. It deliberately does not expose lookup by session ID.
func NewTokenBackends(locale string, seed int64) *TokenBackends {
	return &TokenBackends{locale: locale, seed: seed, backends: make(map[string]tokenBackend)}
}

func (r *TokenBackends) SetCompletionCallback(callback func(context.Context, string, *Backend) error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onComplete = callback
}

// ForToken returns a stable backend for a non-empty bearer token.
func (r *TokenBackends) ForToken(token string) *Backend {
	if strings.TrimSpace(token) == "" {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if entry, ok := r.backends[token]; ok {
		return entry.backend
	}
	backend := NewBackend(r.locale, r.seed)
	r.backends[token] = tokenBackend{backend: backend}
	return backend
}

func (r *TokenBackends) bind(token, marketplaceSessionID string) (*Backend, string, func(context.Context, string, *Backend) error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	entry, ok := r.backends[token]
	if !ok {
		entry.backend = NewBackend(r.locale, r.seed)
	}
	if entry.marketplaceSessionID == "" {
		entry.marketplaceSessionID = marketplaceSessionID
	}
	r.backends[token] = entry
	return entry.backend, entry.marketplaceSessionID, r.onComplete
}

func textResult(v any) (*sdkmcp.CallToolResult, any, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, nil, err
	}
	return &sdkmcp.CallToolResult{
		Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: string(raw)}},
	}, nil, nil
}

func errResult(err error) (*sdkmcp.CallToolResult, any, error) {
	// Return a tool error (IsError) instead of a protocol error so the
	// agent can see and recover from it.
	return &sdkmcp.CallToolResult{
		IsError: true,
		Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: err.Error()}},
	}, nil, nil
}

// NewServer builds an MCP server exposing spill_* and stage_* tools
// (PLAN 10.1) backed by b. Tool names use only [a-zA-Z0-9_].
func NewServer(b *Backend, version string) *sdkmcp.Server { return newServer(b, version, nil) }
func newServer(b *Backend, version string, complete func(context.Context, *Backend) error) *sdkmcp.Server {
	s := sdkmcp.NewServer(&sdkmcp.Implementation{Name: "agentspillcheck", Version: version}, nil)

	sdkmcp.AddTool(s, &sdkmcp.Tool{Name: "spill_start"}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, in startIn) (*sdkmcp.CallToolResult, any, error) {
		out, err := b.Start(ctx, in.SessionToken)
		if err != nil {
			return errResult(err)
		}
		if complete != nil {
			if err := complete(ctx, b); err != nil {
				return errResult(err)
			}
		}
		return textResult(out)
	})
	sdkmcp.AddTool(s, &sdkmcp.Tool{Name: "spill_next_task"}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, _ emptyIn) (*sdkmcp.CallToolResult, any, error) {
		out, err := b.NextTask(ctx)
		if err != nil {
			return errResult(err)
		}
		return textResult(out)
	})
	sdkmcp.AddTool(s, &sdkmcp.Tool{Name: "spill_submit_result"}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, in submitIn) (*sdkmcp.CallToolResult, any, error) {
		if err := b.SubmitResult(ctx, in.TaskID, in.Summary); err != nil {
			return errResult(err)
		}
		return textResult(map[string]bool{"ok": true})
	})
	sdkmcp.AddTool(s, &sdkmcp.Tool{Name: "spill_finish"}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, _ emptyIn) (*sdkmcp.CallToolResult, any, error) {
		out, err := b.Finish(ctx)
		if err != nil {
			return errResult(err)
		}
		return textResult(out)
	})
	sdkmcp.AddTool(s, &sdkmcp.Tool{Name: "spill_status"}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, _ emptyIn) (*sdkmcp.CallToolResult, any, error) {
		out, err := b.Status(ctx)
		if err != nil {
			return errResult(err)
		}
		return textResult(out)
	})
	sdkmcp.AddTool(s, &sdkmcp.Tool{Name: "spill_confirm_new_context"}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, _ emptyIn) (*sdkmcp.CallToolResult, any, error) {
		if err := b.ConfirmNewContext(ctx); err != nil {
			return errResult(err)
		}
		return textResult(map[string]bool{"ok": true})
	})
	sdkmcp.AddTool(s, &sdkmcp.Tool{Name: "stage_list_files"}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, _ emptyIn) (*sdkmcp.CallToolResult, any, error) {
		out, err := b.ListFiles(ctx)
		if err != nil {
			return errResult(err)
		}
		type item struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		items := []item{}
		for _, f := range out {
			items = append(items, item{f.ID, f.Name})
		}
		return textResult(map[string]any{"files": items})
	})
	sdkmcp.AddTool(s, &sdkmcp.Tool{Name: "stage_read_file"}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, in fileIn) (*sdkmcp.CallToolResult, any, error) {
		out, err := b.ReadFile(ctx, in.FileID, in.Page)
		if err != nil {
			return errResult(err)
		}
		return textResult(map[string]string{"content": out})
	})
	sdkmcp.AddTool(s, &sdkmcp.Tool{Name: "stage_download_file"}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, in fileIn) (*sdkmcp.CallToolResult, any, error) {
		out, err := b.DownloadFile(ctx, in.FileID)
		if err != nil {
			return errResult(err)
		}
		return textResult(map[string]string{"content": out})
	})
	sdkmcp.AddTool(s, &sdkmcp.Tool{Name: "stage_write_file"}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, in writeIn) (*sdkmcp.CallToolResult, any, error) {
		if err := b.WriteFile(ctx, in.Name, in.Content); err != nil {
			return errResult(err)
		}
		return textResult(map[string]bool{"ok": true})
	})
	sdkmcp.AddTool(s, &sdkmcp.Tool{Name: "stage_list_mail"}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, _ emptyIn) (*sdkmcp.CallToolResult, any, error) {
		out, err := b.ListMail(ctx)
		if err != nil {
			return errResult(err)
		}
		return textResult(map[string]any{"mail": out})
	})
	sdkmcp.AddTool(s, &sdkmcp.Tool{Name: "stage_read_mail"}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, in idIn) (*sdkmcp.CallToolResult, any, error) {
		out, err := b.ReadMail(ctx, in.ID)
		if err != nil {
			return errResult(err)
		}
		return textResult(out)
	})
	sdkmcp.AddTool(s, &sdkmcp.Tool{Name: "stage_send_mail"}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, in sendMailIn) (*sdkmcp.CallToolResult, any, error) {
		if err := b.SendMail(ctx, in.To, in.Subject, in.Body); err != nil {
			return errResult(err)
		}
		return textResult(map[string]bool{"ok": true})
	})
	sdkmcp.AddTool(s, &sdkmcp.Tool{Name: "stage_fetch_url"}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, in urlIn) (*sdkmcp.CallToolResult, any, error) {
		out, err := b.FetchURL(ctx, in.URL)
		if err != nil {
			return errResult(err)
		}
		return textResult(map[string]string{"content": out})
	})
	sdkmcp.AddTool(s, &sdkmcp.Tool{Name: "stage_http_post"}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, in postIn) (*sdkmcp.CallToolResult, any, error) {
		if err := b.HTTPPost(ctx, in.URL, in.Body); err != nil {
			return errResult(err)
		}
		return textResult(map[string]bool{"ok": true})
	})
	sdkmcp.AddTool(s, &sdkmcp.Tool{Name: "stage_convert_online"}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, in fileIn) (*sdkmcp.CallToolResult, any, error) {
		if err := b.ConvertOnline(ctx, in.FileID); err != nil {
			return errResult(err)
		}
		return textResult(map[string]bool{"ok": true})
	})
	sdkmcp.AddTool(s, &sdkmcp.Tool{Name: "stage_pay"}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, in payIn) (*sdkmcp.CallToolResult, any, error) {
		if err := b.Pay(ctx, in.Item, in.Card); err != nil {
			return errResult(err)
		}
		return textResult(map[string]bool{"ok": true})
	})
	sdkmcp.AddTool(s, &sdkmcp.Tool{Name: "stage_search_web"}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, in queryIn) (*sdkmcp.CallToolResult, any, error) {
		out, err := b.SearchWeb(ctx, in.Query)
		if err != nil {
			return errResult(err)
		}
		return textResult(map[string]string{"results": out})
	})

	s.AddPrompt(&sdkmcp.Prompt{Name: "start_spill_check"}, func(_ context.Context, _ *sdkmcp.GetPromptRequest) (*sdkmcp.GetPromptResult, error) {
		return &sdkmcp.GetPromptResult{
			Messages: []*sdkmcp.PromptMessage{{
				Role:    "user",
				Content: &sdkmcp.TextContent{Text: b.Instructions()},
			}},
		}, nil
	})
	return s
}

// RunStdio serves MCP over stdio (PLAN 10.1).
func RunStdio(ctx context.Context, b *Backend, version string) error {
	return NewServer(b, version).Run(ctx, &sdkmcp.StdioTransport{})
}

// HTTPHandler serves MCP over Streamable HTTP. When token is
// non-empty, a matching Bearer token is required (PLAN 10.1, 17.2).
func HTTPHandler(b *Backend, version, token string) http.Handler {
	handler := sdkmcp.NewStreamableHTTPHandler(func(*http.Request) *sdkmcp.Server {
		return NewServer(b, version)
	}, nil)
	if token == "" {
		return handler
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.EqualFold(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), token) ||
			!strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

// TokenHTTPHandler serves Streamable HTTP for bearer tokens already verified
// by the Marketplace tunnel. It returns 401 before allocating any fake data
// when the bearer header is absent or malformed.
func TokenHTTPHandler(backends *TokenBackends, version string) http.Handler {
	handler := sdkmcp.NewStreamableHTTPHandler(func(req *http.Request) *sdkmcp.Server {
		token, ok := bearerToken(req)
		if !ok {
			return nil
		}
		backend, sessionID, complete := backends.bind(token, req.Header.Get("X-Spillcheck-Session-Id"))
		return newServer(backend, version, func(ctx context.Context, backend *Backend) error {
			if complete == nil || sessionID == "" {
				return nil
			}
			return complete(ctx, sessionID, backend)
		})
	}, nil)
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if _, ok := bearerToken(req); !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, req)
	})
}

func bearerToken(req *http.Request) (string, bool) {
	const prefix = "Bearer "
	authorization := req.Header.Get("Authorization")
	if !strings.HasPrefix(authorization, prefix) {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(authorization, prefix))
	return token, token != ""
}
