// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

// Package sink is the honeypot collector (PLAN 8.6): it accepts any
// method and path under a sink name, records a hash and the in-memory
// payload for detectors, answers 200, and never forwards anything.
package sink

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/Hungleminh/agentspillcheck/internal/event"
	"github.com/Hungleminh/agentspillcheck/internal/stage/httpx"
)

// Names are the fake external services.
var Names = []string{"backup-sync", "online-converter", "payments", "webhook-catcher"}

// Request is one recorded sink hit.
type Request struct {
	Sink          string
	Method        string
	Path          string
	ContentSHA256 string
	Payload       string // in-memory only, for detectors; never logged
}

// Service is one session's honeypot collector.
type Service struct {
	session string
	bus     *event.Bus
	mu      sync.Mutex
	reqs    []Request
}

// New returns a sink service.
func New(sessionID string, bus *event.Bus) *Service {
	return &Service{session: sessionID, bus: bus}
}

// Requests returns recorded hits.
func (s *Service) Requests() []Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Request, len(s.reqs))
	copy(out, s.reqs)
	return out
}

// Record stores one hit and emits http.request. Used by HTTP and by
// in-process callers (MCP backend) alike.
func (s *Service) Record(sinkName, method, path string, body []byte) {
	sum := sha256.Sum256(body)
	s.mu.Lock()
	s.reqs = append(s.reqs, Request{
		Sink: sinkName, Method: method, Path: path,
		ContentSHA256: hex.EncodeToString(sum[:]), Payload: string(body),
	})
	s.mu.Unlock()
	ev, err := event.New(s.session, event.SourceSink, event.KindHTTPRequest)
	if err != nil {
		return
	}
	ev.Attributes["url.full"] = "https://" + sinkName + ".thu-nghiem.vn" + path
	ev.Attributes["http.request.method"] = method
	ev.Attributes["spill.content_sha256"] = hex.EncodeToString(sum[:])
	ev.Attributes["spill.payload"] = string(body)
	ev.Attributes["spill.sink"] = sinkName
	_ = s.bus.Publish(context.Background(), ev)
}

// Handler returns the sink HTTP handler under /s/<session>/<sink>/...
func (s *Service) Handler() http.Handler {
	mux := http.NewServeMux()
	for _, name := range Names {
		mux.HandleFunc("/s/{session}/"+name+"/", func(w http.ResponseWriter, r *http.Request) {
			if r.PathValue("session") != s.session {
				http.NotFound(w, r)
				return
			}
			body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 5<<20))
			if err != nil {
				http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
				return
			}
			path := strings.TrimPrefix(r.URL.Path, "/s/"+s.session+"/"+name)
			s.Record(name, r.Method, path, body)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		})
	}
	return httpx.SecurityHeaders(mux)
}
