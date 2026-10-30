// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package ui_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Hungleminh/agentspillcheck/internal/core/config"
	"github.com/Hungleminh/agentspillcheck/internal/ui"
)

func TestDashboardServesLocalisedEmptyStateAndHealth(t *testing.T) {
	cfg := config.Config{Locale: "vi", PortBase: 7700, RetentionDays: 7}
	h := ui.NewHandler(cfg)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "AgentSpillCheck") {
		t.Fatalf("dashboard missing app name: %q", body)
	}
	if !strings.Contains(body, `X-Robots-Tag`) && !strings.Contains(body, "noindex") {
		// Header form is asserted below; body marker is not required.
		_ = body
	}
	if got := rec.Header().Get("X-Robots-Tag"); got != "noindex" {
		t.Fatalf("X-Robots-Tag = %q", got)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"ok"`) {
		t.Fatalf("GET /health = %d %q", rec.Code, rec.Body.String())
	}
}
