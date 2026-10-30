// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

// Package webmail is the fake webmail stage service (PLAN 8.2).
// Outgoing mail is never delivered; it is kept in a local sink.
// Mail content from agents is untrusted: HTML is escaped by templates
// and request bodies are size-limited (PLAN 17.2).
package webmail

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sync"

	"github.com/Hungleminh/agentspillcheck/internal/core/i18n"
	"github.com/Hungleminh/agentspillcheck/internal/event"
	"github.com/Hungleminh/agentspillcheck/internal/stage/httpx"
)

// Mail is one fake message.
type Mail struct {
	ID      string   `json:"id"`
	From    string   `json:"from"`
	To      string   `json:"to"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
	BaitIDs []string `json:"-"`
}

// Service serves one session's fake mailbox.
type Service struct {
	session string
	bus     *event.Bus
	locale  string
	mu      sync.RWMutex
	mails   []*Mail
	byID    map[string]*Mail
	sent    []*Mail
	seq     int
}

// New returns a webmail service for a session.
func New(sessionID string, bus *event.Bus, locale string) *Service {
	return &Service{session: sessionID, bus: bus, locale: locale, byID: map[string]*Mail{}}
}

// Seed adds a message to the inbox.
func (s *Service) Seed(m Mail) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := m
	s.mails = append(s.mails, &cp)
	s.byID[m.ID] = &cp
}

// Sent returns mail the agent sent (kept in the sink, never delivered).
func (s *Service) Sent() []*Mail {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Mail, len(s.sent))
	copy(out, s.sent)
	return out
}

func (s *Service) emit(kind event.Kind, attrs map[string]any) {
	ev, err := event.New(s.session, event.SourceWebmail, kind)
	if err != nil {
		return
	}
	for k, v := range attrs {
		ev.Attributes[k] = v
	}
	_ = s.bus.Publish(context.Background(), ev)
}

func (s *Service) guard(w http.ResponseWriter, r *http.Request) bool {
	if r.PathValue("session") != s.session {
		http.NotFound(w, r)
		return false
	}
	return true
}

var pageTmpl = template.Must(template.New("mail").Parse(
	`<!DOCTYPE html><html lang="{{.Lang}}"><head><meta charset="utf-8"><title>{{.Title}}</title></head><body><h1>{{.Title}}</h1><ul>{{range .Mail}}<li>{{.Subject}} — {{.From}}</li>{{end}}</ul></body></html>`))

// Handler returns the webmail HTTP handler (API + HTML under /s/<session>/).
func (s *Service) Handler() http.Handler {
	tr, err := i18n.New(s.locale)
	if err != nil {
		tr, _ = i18n.New("en")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /s/{session}/api/mail", func(w http.ResponseWriter, r *http.Request) {
		if !s.guard(w, r) {
			return
		}
		s.mu.RLock()
		mails := make([]*Mail, len(s.mails))
		copy(mails, s.mails)
		s.mu.RUnlock()
		for _, m := range mails {
			s.emit(event.KindCarrierListed, map[string]any{"spill.carrier_id": m.ID, "spill.bait_ids": m.BaitIDs})
		}
		writeJSON(w, http.StatusOK, map[string]any{"mail": mails})
	})
	mux.HandleFunc("GET /s/{session}/api/mail/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !s.guard(w, r) {
			return
		}
		s.mu.RLock()
		m, ok := s.byID[r.PathValue("id")]
		s.mu.RUnlock()
		if !ok {
			http.NotFound(w, r)
			return
		}
		s.emit(event.KindCarrierOpened, map[string]any{"spill.carrier_id": m.ID, "spill.bait_ids": m.BaitIDs})
		writeJSON(w, http.StatusOK, m)
	})
	mux.HandleFunc("POST /s/{session}/api/mail/send", func(w http.ResponseWriter, r *http.Request) {
		if !s.guard(w, r) {
			return
		}
		httpx.MaxBody(w, r, 1<<20)
		var in Mail
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			// MaxBytesReader errors mean the body exceeded the limit.
			http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
			return
		}
		if in.To == "" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		s.mu.Lock()
		s.seq++
		in.ID = fmt.Sprintf("sent-%d", s.seq)
		stored := in
		s.sent = append(s.sent, &stored)
		s.mu.Unlock()
		sum := sha256.Sum256([]byte(in.Body))
		s.emit(event.KindMailSent, map[string]any{
			"spill.to": in.To, "spill.content_sha256": hex.EncodeToString(sum[:])})
		writeJSON(w, http.StatusCreated, map[string]any{"id": stored.ID, "to": in.To})
	})
	mux.HandleFunc("GET /s/{session}/", func(w http.ResponseWriter, r *http.Request) {
		if !s.guard(w, r) {
			return
		}
		s.mu.RLock()
		mails := make([]*Mail, len(s.mails))
		copy(mails, s.mails)
		s.mu.RUnlock()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = pageTmpl.Execute(w, map[string]any{"Lang": tr.Locale(), "Title": tr.T("mail.title"), "Mail": mails})
	})
	return httpx.SecurityHeaders(mux)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
