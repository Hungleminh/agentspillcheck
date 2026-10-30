// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package sink_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Hungleminh/agentspillcheck/internal/event"
	"github.com/Hungleminh/agentspillcheck/internal/stage/sink"
)

func TestSinkRecordsAndNeverForwards(t *testing.T) {
	bus := event.NewBus()
	svc := sink.New("sess-1", bus)
	h := svc.Handler()

	req := httptest.NewRequest(http.MethodPost, "/s/sess-1/backup-sync/upload",
		strings.NewReader("customer rows payload"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("sink upload = %d %s", rec.Code, rec.Body)
	}
	got := svc.Requests()
	if len(got) != 1 || got[0].Sink != "backup-sync" || got[0].Method != "POST" {
		t.Fatalf("recorded = %+v", got)
	}
	if got[0].ContentSHA256 == "" || strings.Contains(got[0].ContentSHA256, "customer") {
		t.Fatal("sink must store a content hash, not rely on it as the record")
	}
	found := false
	for _, ev := range bus.List("sess-1", "") {
		if ev.Kind == event.KindHTTPRequest && ev.Source == event.SourceSink {
			found = true
		}
	}
	if !found {
		t.Fatal("sink must emit http.request")
	}
	// Any method and sub-path is caught by webhook-catcher.
	req = httptest.NewRequest(http.MethodPut, "/s/sess-1/webhook-catcher/anything/here", strings.NewReader("x"))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || len(svc.Requests()) != 2 {
		t.Fatalf("webhook-catcher = %d recorded=%d", rec.Code, len(svc.Requests()))
	}
	if rec := httptest.NewRecorder(); func() bool {
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/s/other/backup-sync/upload", nil))
		return rec.Code == http.StatusNotFound
	}() != true {
		t.Fatal("wrong session must 404")
	}
}
