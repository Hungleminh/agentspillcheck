// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

// Package api is the REST API (PLAN 15). A Bearer token is required
// for everything except /health when a token is configured (public
// mode); locally with no token it is open on 127.0.0.1 only because
// the server binds there.
package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Hungleminh/agentspillcheck/internal/core/session"
	"github.com/Hungleminh/agentspillcheck/internal/report"
	"github.com/Hungleminh/agentspillcheck/internal/stage/httpx"
)

// NewHandler returns the /api/v1 handler.
func NewHandler(mgr *session.Manager, version, token string) http.Handler {
	mux := http.NewServeMux()
	write := func(w http.ResponseWriter, code int, v any) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(v)
	}
	lookup := func(w http.ResponseWriter, r *http.Request) (string, bool) {
		id := r.PathValue("id")
		if _, ok := mgr.Get(id); !ok {
			http.NotFound(w, r)
			return "", false
		}
		return id, true
	}

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, _ *http.Request) {
		write(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /api/v1/version", func(w http.ResponseWriter, _ *http.Request) {
		write(w, http.StatusOK, map[string]string{"version": version})
	})
	mux.HandleFunc("POST /api/v1/sessions", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Tier       string `json:"tier"`
			Mode       string `json:"mode"`
			Locale     string `json:"locale"`
			AgentLabel string `json:"agent_label"`
		}
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&in)
		}
		if in.Locale == "" {
			in.Locale = "en"
		}
		info, err := mgr.Create(in.Locale, in.Tier, in.Mode, in.AgentLabel)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		write(w, http.StatusCreated, map[string]any{
			"session_id": info.SessionID, "persona_summary": info.PersonaSummary, "session": info})
	})
	mux.HandleFunc("GET /api/v1/sessions", func(w http.ResponseWriter, _ *http.Request) {
		write(w, http.StatusOK, map[string]any{"sessions": mgr.List()})
	})
	mux.HandleFunc("GET /api/v1/sessions/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, ok := lookup(w, r)
		if !ok {
			return
		}
		info, _ := mgr.Get(id)
		write(w, http.StatusOK, info)
	})
	mux.HandleFunc("POST /api/v1/sessions/{id}/tasks/next", func(w http.ResponseWriter, r *http.Request) {
		id, ok := lookup(w, r)
		if !ok {
			return
		}
		task, err := mgr.NextTask(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		write(w, http.StatusOK, task)
	})
	mux.HandleFunc("POST /api/v1/sessions/{id}/tasks/{task_id}/result", func(w http.ResponseWriter, r *http.Request) {
		id, ok := lookup(w, r)
		if !ok {
			return
		}
		var in struct {
			Summary string `json:"summary"`
			Output  string `json:"output"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		summary := in.Summary
		if summary == "" {
			summary = in.Output
		}
		if err := mgr.SubmitResult(id, r.PathValue("task_id"), summary); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		write(w, http.StatusOK, map[string]bool{"ok": true})
	})
	mux.HandleFunc("POST /api/v1/sessions/{id}/finish", func(w http.ResponseWriter, r *http.Request) {
		id, ok := lookup(w, r)
		if !ok {
			return
		}
		out, err := mgr.Finish(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		write(w, http.StatusOK, out)
	})
	mux.HandleFunc("GET /api/v1/sessions/{id}/events", func(w http.ResponseWriter, r *http.Request) {
		id, ok := lookup(w, r)
		if !ok {
			return
		}
		events, _ := mgr.Events(id)
		w.Header().Set("Content-Type", "application/x-ndjson")
		enc := json.NewEncoder(w)
		for _, ev := range events {
			_ = enc.Encode(ev)
		}
	})
	reportHandler := func(format string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id, ok := lookup(w, r)
			if !ok {
				return
			}
			rep, _ := mgr.Report(id)
			locale := r.URL.Query().Get("locale")
			if locale == "" {
				locale = "en"
			}
			switch format {
			case "json":
				write(w, http.StatusOK, rep)
			case "html":
				out, err := report.RenderHTML(rep, locale)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				_, _ = w.Write([]byte(out))
			case "md":
				out, err := report.RenderMarkdown(rep, locale)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
				_, _ = w.Write([]byte(out))
			}
		}
	}
	mux.HandleFunc("GET /api/v1/sessions/{id}/report.json", reportHandler("json"))
	mux.HandleFunc("GET /api/v1/sessions/{id}/report.html", reportHandler("html"))
	mux.HandleFunc("GET /api/v1/sessions/{id}/report.md", reportHandler("md"))

	handler := httpx.SecurityHeaders(mux)
	if token == "" {
		return handler
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/health" {
			handler.ServeHTTP(w, r)
			return
		}
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") || strings.TrimPrefix(auth, "Bearer ") != token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, r)
	})
}
