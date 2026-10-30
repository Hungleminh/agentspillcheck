// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

// Package drive is the fake drive stage service (PLAN 8.3): files are
// served page by page, every access is recorded as an event, and agent
// writes are kept and recorded. File IDs are opaque; no filesystem
// path is ever derived from agent input.
package drive

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"sync"

	"github.com/Hungleminh/agentspillcheck/internal/core/i18n"
	"github.com/Hungleminh/agentspillcheck/internal/event"
	"github.com/Hungleminh/agentspillcheck/internal/stage/httpx"
)

// PageBytes is the page size for text-like carriers (PLAN 8.3: 2 KB).
const PageBytes = 2048

// File is one seeded carrier in the drive.
type File struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Content string   `json:"-"`
	BaitIDs []string `json:"-"`
}

// Service serves one session's fake drive.
type Service struct {
	session string
	bus     *event.Bus
	locale  string
	mu      sync.RWMutex
	files   []*File
	byID    map[string]*File
	seq     int
}

// New returns a drive service for a session.
func New(sessionID string, bus *event.Bus, locale string) *Service {
	return &Service{session: sessionID, bus: bus, locale: locale, byID: map[string]*File{}}
}

// Seed adds a file to the drive.
func (s *Service) Seed(f File) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := f
	s.files = append(s.files, &cp)
	s.byID[f.ID] = &cp
}

func (s *Service) emit(kind event.Kind, attrs map[string]any) {
	ev, err := event.New(s.session, event.SourceDrive, kind)
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

var pageTmpl = template.Must(template.New("drive").Parse(
	`<!DOCTYPE html><html lang="{{.Lang}}"><head><meta charset="utf-8"><title>{{.Title}}</title></head><body><h1>{{.Title}}</h1><ul>{{range .Files}}<li>{{.Name}}</li>{{end}}</ul></body></html>`))

// Handler returns the drive HTTP handler (API + HTML under /s/<session>/).
func (s *Service) Handler() http.Handler {
	tr, err := i18n.New(s.locale)
	if err != nil {
		tr, _ = i18n.New("en")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /s/{session}/api/files", func(w http.ResponseWriter, r *http.Request) {
		if !s.guard(w, r) {
			return
		}
		s.mu.RLock()
		files := make([]*File, len(s.files))
		copy(files, s.files)
		s.mu.RUnlock()
		for _, f := range files {
			s.emit(event.KindCarrierListed, map[string]any{"spill.carrier_id": f.ID, "spill.bait_ids": f.BaitIDs})
		}
		type item struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		items := []item{}
		for _, f := range files {
			items = append(items, item{f.ID, f.Name})
		}
		writeJSON(w, http.StatusOK, map[string]any{"files": items})
	})
	mux.HandleFunc("GET /s/{session}/api/files/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !s.guard(w, r) {
			return
		}
		f, ok := s.lookup(r.PathValue("id"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		pages := splitPages(f.Content)
		page := 1
		if raw := r.URL.Query().Get("page"); raw != "" {
			if n, err := strconv.Atoi(raw); err == nil && n >= 1 {
				page = n
			}
		}
		if page > len(pages) {
			page = len(pages)
		}
		s.emit(event.KindCarrierPageRead, map[string]any{
			"spill.carrier_id": f.ID, "spill.bait_ids": f.BaitIDs, "spill.page": page})
		writeJSON(w, http.StatusOK, map[string]any{
			"id": f.ID, "name": f.Name, "page": page, "pages": len(pages), "content": pages[page-1]})
	})
	mux.HandleFunc("GET /s/{session}/api/files/{id}/download", func(w http.ResponseWriter, r *http.Request) {
		if !s.guard(w, r) {
			return
		}
		f, ok := s.lookup(r.PathValue("id"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		s.emit(event.KindCarrierDownloaded, map[string]any{"spill.carrier_id": f.ID, "spill.bait_ids": f.BaitIDs})
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write([]byte(f.Content))
	})
	mux.HandleFunc("POST /s/{session}/api/files", func(w http.ResponseWriter, r *http.Request) {
		if !s.guard(w, r) {
			return
		}
		httpx.MaxBody(w, r, 1<<20)
		var in struct {
			Name    string `json:"name"`
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Name == "" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		s.mu.Lock()
		s.seq++
		f := &File{ID: fmt.Sprintf("agent-file-%d", s.seq), Name: in.Name, Content: in.Content}
		s.files = append(s.files, f)
		s.byID[f.ID] = f
		s.mu.Unlock()
		s.emit(event.KindHTTPRequest, map[string]any{
			"http.request.method": "POST", "spill.carrier_id": f.ID})
		writeJSON(w, http.StatusCreated, map[string]any{"id": f.ID, "name": f.Name})
	})
	mux.HandleFunc("GET /s/{session}/", func(w http.ResponseWriter, r *http.Request) {
		if !s.guard(w, r) {
			return
		}
		s.mu.RLock()
		files := make([]*File, len(s.files))
		copy(files, s.files)
		s.mu.RUnlock()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = pageTmpl.Execute(w, map[string]any{"Lang": tr.Locale(), "Title": tr.T("drive.title"), "Files": files})
	})
	return httpx.SecurityHeaders(mux)
}

// Files lists the seeded files and records carrier.listed events.
func (s *Service) Files() []*File {
	s.mu.RLock()
	files := make([]*File, len(s.files))
	copy(files, s.files)
	s.mu.RUnlock()
	for _, f := range files {
		s.emit(event.KindCarrierListed, map[string]any{"spill.carrier_id": f.ID, "spill.bait_ids": f.BaitIDs})
	}
	return files
}

// ReadPage returns one page of a file and records carrier.page_read.
func (s *Service) ReadPage(id string, page int) (string, int, error) {
	f, ok := s.lookup(id)
	if !ok {
		return "", 0, fmt.Errorf("drive: unknown file %q", id)
	}
	pages := splitPages(f.Content)
	if page < 1 {
		page = 1
	}
	if page > len(pages) {
		page = len(pages)
	}
	s.emit(event.KindCarrierPageRead, map[string]any{
		"spill.carrier_id": f.ID, "spill.bait_ids": f.BaitIDs, "spill.page": page})
	return pages[page-1], len(pages), nil
}

// Download returns the full content and records carrier.downloaded.
func (s *Service) Download(id string) (string, error) {
	f, ok := s.lookup(id)
	if !ok {
		return "", fmt.Errorf("drive: unknown file %q", id)
	}
	s.emit(event.KindCarrierDownloaded, map[string]any{"spill.carrier_id": f.ID, "spill.bait_ids": f.BaitIDs})
	return f.Content, nil
}

// Write stores a file written by the agent and records it.
func (s *Service) Write(name, content string) (*File, error) {
	if name == "" {
		return nil, fmt.Errorf("drive: name is required")
	}
	s.mu.Lock()
	s.seq++
	f := &File{ID: fmt.Sprintf("agent-file-%d", s.seq), Name: name, Content: content}
	s.files = append(s.files, f)
	s.byID[f.ID] = f
	s.mu.Unlock()
	s.emit(event.KindHTTPRequest, map[string]any{"http.request.method": "POST", "spill.carrier_id": f.ID})
	return f, nil
}

func (s *Service) lookup(id string) (*File, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f, ok := s.byID[id]
	return f, ok
}

func splitPages(content string) []string {
	if content == "" {
		return []string{""}
	}
	var pages []string
	for i := 0; i < len(content); i += PageBytes {
		end := i + PageBytes
		if end > len(content) {
			end = len(content)
		}
		pages = append(pages, content[i:end])
	}
	return pages
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
