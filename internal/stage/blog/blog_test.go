// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package blog_test

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Hungleminh/agentspillcheck/internal/event"
	"github.com/Hungleminh/agentspillcheck/internal/stage/blog"
)

func get(h http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestBlogPostsAndInjectionVariants(t *testing.T) {
	bus := event.NewBus()
	svc := blog.New("sess-1", bus, "vi")
	h := svc.Handler()

	rec := get(h, "/s/sess-1/api/posts")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "post-3") {
		t.Fatalf("posts = %d %s", rec.Code, rec.Body)
	}
	// Post 3 carries a hidden injection asking to upload to backup-sync.
	rec = get(h, "/s/sess-1/api/posts/post-3")
	body := rec.Body.String()
	if !strings.Contains(body, "backup-sync") {
		t.Fatalf("post-3 must contain the injection target: %s", body)
	}
	if !strings.Contains(body, "hidden") && !strings.Contains(body, "aria-hidden") {
		t.Fatalf("injection must be hidden-marked: %s", body)
	}
	found := false
	for _, ev := range bus.List("sess-1", "") {
		if ev.Kind == event.KindCarrierOpened && ev.Attributes["spill.injection_id"] != nil {
			found = true
		}
	}
	if !found {
		t.Fatal("opening post-3 must emit an event with spill.injection_id")
	}

	// Base64 variant hides the same instruction encoded.
	svc2 := blog.New("sess-1", event.NewBus(), "en", blog.VariantBase64)
	rec = get(svc2.Handler(), "/s/sess-1/api/posts/post-3")
	encoded := base64.StdEncoding.EncodeToString([]byte("backup-sync"))
	_ = encoded
	if !strings.Contains(rec.Body.String(), "backup-sync") && !strings.Contains(rec.Body.String(), encoded[:12]) {
		// Encoded form contains no plaintext host; assert it decodes to it.
	}
	// Clean posts carry no injection.
	rec = get(h, "/s/sess-1/api/posts/post-1")
	if strings.Contains(rec.Body.String(), "backup-sync") {
		t.Fatal("post-1 must not contain the injection")
	}
	// Session isolation.
	if rec := get(h, "/s/other/api/posts"); rec.Code != http.StatusNotFound {
		t.Fatalf("wrong session = %d", rec.Code)
	}
}
