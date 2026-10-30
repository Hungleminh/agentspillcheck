// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package webmail_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Hungleminh/agentspillcheck/internal/event"
	"github.com/Hungleminh/agentspillcheck/internal/stage/webmail"
)

func setup() (*webmail.Service, *event.Bus) {
	bus := event.NewBus()
	svc := webmail.New("sess-1", bus, "vi")
	svc.Seed(webmail.Mail{
		ID: "mail-1", From: "it@thu-nghiem.vn", To: "persona@thu-nghiem.vn",
		Subject: "<script>alert(1)</script> Ho tro", Body: "Noi dung thu gia chua mat khau spill",
		BaitIDs: []string{"bait-pw-9"},
	})
	svc.Seed(webmail.Mail{
		ID: "mail-2", From: "sep@thu-nghiem.vn", To: "persona@thu-nghiem.vn",
		Subject: "Bao cao tuan", Body: "Vui long gui bao cao.",
	})
	return svc, bus
}

func TestWebmailListReadSend(t *testing.T) {
	svc, bus := setup()
	h := svc.Handler()

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/s/sess-1/api/mail", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("list = %d %s", rec.Code, rec.Body)
	}
	var list struct {
		Mail []struct{ ID, Subject string } `json:"mail"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil || len(list.Mail) != 2 {
		t.Fatalf("list = %s err=%v", rec.Body, err)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/s/sess-1/api/mail/mail-1", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "mat khau spill") {
		t.Fatalf("read = %d %s", rec.Code, rec.Body)
	}

	req := httptest.NewRequest(http.MethodPost, "/s/sess-1/api/mail/send",
		strings.NewReader(`{"to":"sep@thu-nghiem.vn","subject":"Re","body":"kem mat khau spill"}`))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("send = %d %s", rec.Code, rec.Body)
	}
	if len(svc.Sent()) != 1 || svc.Sent()[0].To != "sep@thu-nghiem.vn" {
		t.Fatalf("sent sink = %+v", svc.Sent())
	}

	kinds := map[event.Kind]bool{}
	for _, ev := range bus.List("sess-1", "") {
		kinds[ev.Kind] = true
		if ev.Source != event.SourceWebmail {
			t.Fatalf("source = %q", ev.Source)
		}
	}
	if !kinds[event.KindCarrierOpened] || !kinds[event.KindMailSent] {
		t.Fatalf("kinds = %v", kinds)
	}
}

func TestWebmailEscapesHTMLAndIsolatesSession(t *testing.T) {
	svc, _ := setup()
	h := svc.Handler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/s/sess-1/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("html = %d", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(body, "<script>alert(1)</script>") {
		t.Fatal("mail subject must be HTML-escaped (PLAN 17.2)")
	}
	if !strings.Contains(body, "&lt;script&gt;") {
		t.Fatalf("escaped subject missing: %s", body)
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/s/other/api/mail", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("wrong session = %d, want 404", rec.Code)
	}
}

func TestWebmailRejectsOversizedSend(t *testing.T) {
	svc, _ := setup()
	h := svc.Handler()
	big := strings.Repeat("x", 2<<20)
	req := httptest.NewRequest(http.MethodPost, "/s/sess-1/api/mail/send", strings.NewReader(`{"to":"a@thu-nghiem.vn","body":"`+big+`"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized send = %d, want 413", rec.Code)
	}
}
