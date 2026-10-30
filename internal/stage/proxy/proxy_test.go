// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package proxy_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Hungleminh/agentspillcheck/internal/event"
	"github.com/Hungleminh/agentspillcheck/internal/stage/proxy"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestProxyRecordsConnectAndBlocksPrivate(t *testing.T) {
	bus := event.NewBus()
	upstream := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("stage-ok")),
			Request:    r,
		}, nil
	})
	p := proxy.New("sess-1", bus, proxy.Options{
		AllowHosts: map[string]bool{"stage.thu-nghiem.vn:80": true},
		Transport:  upstream,
	})
	h := p.Handler()

	// Absolute-URL proxy request to an allowed stage host.
	req := httptest.NewRequest(http.MethodGet, "http://stage.thu-nghiem.vn:80/hello", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != "stage-ok" {
		t.Fatalf("proxied = %d %q", rec.Code, rec.Body.String())
	}
	found := false
	for _, ev := range bus.List("sess-1", "") {
		if ev.Kind == event.KindProxyConnect && fmt.Sprint(ev.Attributes["server.address"]) == "stage.thu-nghiem.vn:80" {
			found = true
		}
	}
	if !found {
		t.Fatal("proxy must emit proxy.connect with server.address")
	}

	// A private destination that is not an allowed stage host is blocked (SSRF).
	req = httptest.NewRequest(http.MethodGet, "http://127.0.0.1:9/secret", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("private target = %d, want 403", rec.Code)
	}
}

func TestProxyRejectsRelativeURL(t *testing.T) {
	p := proxy.New("sess-1", event.NewBus(), proxy.Options{})
	rec := httptest.NewRecorder()
	p.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/relative", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("relative URL = %d, want 400", rec.Code)
	}
}
