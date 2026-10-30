// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Hungleminh/agentspillcheck/internal/api"
	"github.com/Hungleminh/agentspillcheck/internal/core/session"
	"github.com/Hungleminh/agentspillcheck/internal/store"
)

func setup(t *testing.T, token string) http.Handler {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "data.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	mgr := session.NewManager(st)
	return api.NewHandler(mgr, "v0.0.0-test", token)
}

func do(t *testing.T, h http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestAPIFullFlow(t *testing.T) {
	h := setup(t, "")
	rec := do(t, h, "POST", "/api/v1/sessions", `{"tier":"standard","locale":"vi","agent_label":"Test Agent"}`, "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("create = %d %s", rec.Code, rec.Body)
	}
	var created struct {
		SessionID string `json:"session_id"`
		Persona   string `json:"persona_summary"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil || created.SessionID == "" {
		t.Fatalf("created = %s err=%v", rec.Body, err)
	}
	id := created.SessionID

	rec = do(t, h, "POST", "/api/v1/sessions/"+id+"/tasks/next", "", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "task_id") {
		t.Fatalf("next = %d %s", rec.Code, rec.Body)
	}
	var task struct {
		TaskID string `json:"task_id"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &task)

	rec = do(t, h, "POST", "/api/v1/sessions/"+id+"/tasks/"+task.TaskID+"/result", `{"summary":"done"}`, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("result = %d %s", rec.Code, rec.Body)
	}
	rec = do(t, h, "POST", "/api/v1/sessions/"+id+"/finish", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("finish = %d %s", rec.Code, rec.Body)
	}
	rec = do(t, h, "GET", "/api/v1/sessions/"+id+"/report.json", "", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"grade"`) {
		t.Fatalf("report.json = %d %s", rec.Code, rec.Body)
	}
	rec = do(t, h, "GET", "/api/v1/sessions/"+id+"/report.html", "", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "AgentSpillCheck") {
		t.Fatalf("report.html = %d", rec.Code)
	}
	rec = do(t, h, "GET", "/api/v1/sessions/"+id+"/events", "", "")
	if rec.Code != http.StatusOK || rec.Body.String() == "" {
		t.Fatalf("events = %d", rec.Code)
	}
	rec = do(t, h, "GET", "/api/v1/health", "", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "ok") {
		t.Fatalf("health = %d %s", rec.Code, rec.Body)
	}
	rec = do(t, h, "GET", "/api/v1/sessions/nope", "", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unknown session = %d", rec.Code)
	}
}

func TestAPIRequiresBearerWhenTokenSet(t *testing.T) {
	h := setup(t, "tok123")
	if rec := do(t, h, "GET", "/api/v1/sessions", "", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("no token = %d", rec.Code)
	}
	if rec := do(t, h, "GET", "/api/v1/sessions", "", "wrong"); rec.Code != http.StatusUnauthorized {
		t.Fatalf("wrong token = %d", rec.Code)
	}
	if rec := do(t, h, "GET", "/api/v1/sessions", "", "tok123"); rec.Code != http.StatusOK {
		t.Fatalf("good token = %d", rec.Code)
	}
	// Health stays public.
	if rec := do(t, h, "GET", "/api/v1/health", "", ""); rec.Code != http.StatusOK {
		t.Fatalf("health = %d", rec.Code)
	}
}
