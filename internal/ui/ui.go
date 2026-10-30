// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

// Package ui serves the dashboard. Server-rendered, no build step,
// no external resources (PLAN 8.8, 17.2).
package ui

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/Hungleminh/agentspillcheck/internal/core/config"
	"github.com/Hungleminh/agentspillcheck/internal/core/i18n"
)

var page = template.Must(template.New("dashboard").Parse(`<!DOCTYPE html>
<html lang="{{.Lang}}"><head><meta charset="utf-8">
<meta name="robots" content="noindex">
<title>{{.Name}} — {{.Title}}</title>
<style>body{font-family:system-ui,sans-serif;margin:2rem;max-width:60rem}h1{color:#0E7C6B}</style>
</head><body><h1>{{.Name}}</h1><p>{{.Tagline}}</p>
<h2>{{.Title}}</h2><p>{{.Empty}}</p>
<h3>{{.Sessions}}</h3><p>{{.CreateHint}}</p>
<h3>{{.Reports}}</h3><p><small>{{.Responsible}}</small></p>
</body></html>`))

// NewHandler returns the dashboard handler for cfg.
func NewHandler(cfg config.Config) http.Handler {
	tr, err := i18n.New(cfg.Locale)
	if err != nil {
		tr, _ = i18n.New("en")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = page.Execute(w, map[string]string{
			"Lang": tr.Locale(), "Name": tr.T("app.name"),
			"Tagline": tr.T("app.tagline"), "Title": tr.T("dashboard.title"),
			"Empty": tr.T("dashboard.empty"), "Responsible": tr.T("dashboard.responsible_use"),
			"Sessions": tr.T("dashboard.sessions"), "CreateHint": tr.T("dashboard.create_hint"),
			"Reports": tr.T("dashboard.reports"),
		})
	})
	return securityHeaders(mux)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Robots-Tag", "noindex")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; img-src data:")
		next.ServeHTTP(w, r)
	})
}
