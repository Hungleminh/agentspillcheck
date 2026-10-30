// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

// Package session manages live test sessions: it owns one MCP backend
// per session and persists its events to the store as they appear.
// Persistence stores event attributes as recorded; values flagged as
// payloads exist only to feed detectors in memory, and a persisted
// session keeps the same attributes, so session directories must be
// protected like session data (0700, retention) per PLAN 17.2.
package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Hungleminh/agentspillcheck/internal/analyze"
	"github.com/Hungleminh/agentspillcheck/internal/event"
	spillmcp "github.com/Hungleminh/agentspillcheck/internal/mcp"
	"github.com/Hungleminh/agentspillcheck/internal/report"
	"github.com/Hungleminh/agentspillcheck/internal/store"
)

// Info describes one live session.
type Info struct {
	SessionID      string    `json:"session_id"`
	PersonaSummary string    `json:"persona_summary"`
	Locale         string    `json:"locale"`
	AgentLabel     string    `json:"agent_label,omitempty"`
	Mode           string    `json:"mode,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

type entry struct {
	backend   *spillmcp.Backend
	info      Info
	persisted int
}

// Manager owns all live sessions.
type Manager struct {
	mu       sync.Mutex
	store    *store.Store
	seq      int
	sessions map[string]*entry
}

// NewManager returns a manager persisting to st (nil store disables
// persistence).
func NewManager(st *store.Store) *Manager {
	return &Manager{store: st, sessions: map[string]*entry{}}
}

// Create starts a new session.
func (m *Manager) Create(locale, tier, mode, agentLabel string) (Info, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seq++
	seed := int64(1000 + m.seq)
	backend := spillmcp.NewBackend(locale, seed)
	start, err := backend.Start(context.Background(), "")
	if err != nil {
		return Info{}, err
	}
	info := Info{
		SessionID: start.SessionID, PersonaSummary: start.PersonaSummary,
		Locale: locale, AgentLabel: agentLabel, Mode: mode,
		Status: "running", CreatedAt: time.Now().UTC(),
	}
	_ = tier // tiers expand the task set in a later step
	e := &entry{backend: backend, info: info}
	m.sessions[info.SessionID] = e
	if m.store != nil {
		_ = m.store.CreateSession(context.Background(), store.Session{
			ID: info.SessionID, TenantID: "local", Locale: locale,
			Mode: mode, AgentLabel: agentLabel, Status: "running",
			CreatedAt: info.CreatedAt, ExpiresAt: info.CreatedAt.Add(7 * 24 * time.Hour),
		})
	}
	m.syncLocked(e)
	return info, nil
}

// Get returns one session's info.
func (m *Manager) Get(id string) (Info, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.sessions[id]
	if !ok {
		return Info{}, false
	}
	return e.info, true
}

// List returns all live sessions.
func (m *Manager) List() []Info {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []Info{}
	for _, e := range m.sessions {
		out = append(out, e.info)
	}
	return out
}

// Expire ends and drops a session.
func (m *Manager) Expire(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.sessions[id]; !ok {
		return false
	}
	delete(m.sessions, id)
	return true
}

// Backend returns the session's backend.
func (m *Manager) Backend(id string) (*spillmcp.Backend, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.sessions[id]
	if !ok {
		return nil, false
	}
	return e.backend, true
}

// NextTask delegates to the session backend and persists new events.
func (m *Manager) NextTask(id string) (spillmcp.TaskResult, error) {
	b, ok := m.Backend(id)
	if !ok {
		return spillmcp.TaskResult{}, fmt.Errorf("session: unknown session %q", id)
	}
	task, err := b.NextTask(context.Background())
	m.sync(id)
	return task, err
}

// SubmitResult delegates and persists new events.
func (m *Manager) SubmitResult(id, taskID, summary string) error {
	b, ok := m.Backend(id)
	if !ok {
		return fmt.Errorf("session: unknown session %q", id)
	}
	err := b.SubmitResult(context.Background(), taskID, summary)
	m.sync(id)
	return err
}

// Finish closes a session and persists remaining events.
func (m *Manager) Finish(id string) (spillmcp.FinishResult, error) {
	b, ok := m.Backend(id)
	if !ok {
		return spillmcp.FinishResult{}, fmt.Errorf("session: unknown session %q", id)
	}
	out, err := b.Finish(context.Background())
	m.mu.Lock()
	if e, ok := m.sessions[id]; ok {
		e.info.Status = "finished"
	}
	m.mu.Unlock()
	m.sync(id)
	return out, err
}

// Events returns persisted-order events for a session.
func (m *Manager) Events(id string) ([]event.Event, bool) {
	b, ok := m.Backend(id)
	if !ok {
		return nil, false
	}
	return b.Events(), true
}

// Report builds and analyzes the session report.
func (m *Manager) Report(id string) (report.Report, bool) {
	b, ok := m.Backend(id)
	if !ok {
		return report.Report{}, false
	}
	rep := report.Build(id, b.Baits(), b.Events())
	return analyze.Enrich(rep, b.Events()), true
}

func (m *Manager) sync(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if e, ok := m.sessions[id]; ok {
		m.syncLocked(e)
	}
}

func (m *Manager) syncLocked(e *entry) {
	if m.store == nil {
		return
	}
	events := e.backend.Events()
	for ; e.persisted < len(events); e.persisted++ {
		_ = m.store.AppendEvent(context.Background(), events[e.persisted])
	}
}
