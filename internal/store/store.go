// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

// Package store persists sessions and append-only events in SQLite
// (WAL mode, PLAN 6.3). Events and sessions never contain full bait
// values; events reference baits by spill.bait_ids only.
package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite" // pure-Go SQLite driver

	"github.com/Hungleminh/agentspillcheck/internal/event"
)

// Session is the stored session record (PLAN 6.1).
type Session struct {
	ID            string
	TenantID      string
	CreatedAt     time.Time
	ExpiresAt     time.Time
	Mode          string
	ConnectorName string
	AgentLabel    string
	Locale        string
	Status        string
}

// Store is a SQLite-backed store.
type Store struct {
	db *sql.DB
}

// Open opens (creating if needed) the database at path and migrates it.
func Open(path string) (*Store, error) {
	if path == "" {
		return nil, fmt.Errorf("store: path is required")
	}
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) migrate(ctx context.Context) error {
	stmts := []string{
		`PRAGMA journal_mode=WAL`,
		`PRAGMA foreign_keys=ON`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL DEFAULT 'local',
			created_at TEXT NOT NULL, expires_at TEXT, mode TEXT,
			connector_name TEXT, agent_label TEXT, locale TEXT, status TEXT)`,
		`CREATE TABLE IF NOT EXISTS events (
			event_id TEXT PRIMARY KEY, session_id TEXT NOT NULL,
			ts TEXT NOT NULL, source TEXT NOT NULL, kind TEXT NOT NULL,
			attributes TEXT NOT NULL DEFAULT '{}')`,
		`CREATE INDEX IF NOT EXISTS idx_events_session ON events(session_id)`,
	}
	for _, q := range stmts {
		if _, err := s.db.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("store: migrate: %w", err)
		}
	}
	return nil
}

// Close closes the database.
func (s *Store) Close() error { return s.db.Close() }

// JournalMode returns the current journal mode (tests and doctor).
func (s *Store) JournalMode(ctx context.Context) (string, error) {
	var mode string
	err := s.db.QueryRowContext(ctx, `PRAGMA journal_mode`).Scan(&mode)
	return mode, err
}

// CreateSession inserts a session.
func (s *Store) CreateSession(ctx context.Context, sess Session) error {
	if sess.TenantID == "" {
		sess.TenantID = "local"
	}
	created := sess.CreatedAt
	if created.IsZero() {
		created = time.Now().UTC()
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO sessions (id, tenant_id, created_at, expires_at, mode, connector_name, agent_label, locale, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sess.ID, sess.TenantID, created.Format(time.RFC3339Nano),
		sess.ExpiresAt.Format(time.RFC3339Nano), sess.Mode, sess.ConnectorName,
		sess.AgentLabel, sess.Locale, sess.Status)
	return err
}

// GetSession loads a session by ID.
func (s *Store) GetSession(ctx context.Context, id string) (Session, error) {
	var sess Session
	var created, expires string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, created_at, expires_at, mode, connector_name, agent_label, locale, status
		 FROM sessions WHERE id = ?`, id).
		Scan(&sess.ID, &sess.TenantID, &created, &expires, &sess.Mode,
			&sess.ConnectorName, &sess.AgentLabel, &sess.Locale, &sess.Status)
	if err != nil {
		return Session{}, err
	}
	sess.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	sess.ExpiresAt, _ = time.Parse(time.RFC3339Nano, expires)
	return sess, nil
}

// AppendEvent appends one event (append-only; no update path exists).
func (s *Store) AppendEvent(ctx context.Context, ev event.Event) error {
	attrs, err := json.Marshal(ev.Attributes)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO events (event_id, session_id, ts, source, kind, attributes) VALUES (?, ?, ?, ?, ?, ?)`,
		ev.EventID, ev.SessionID, ev.TS.Format(time.RFC3339Nano),
		string(ev.Source), string(ev.Kind), string(attrs))
	return err
}

// ListEvents returns events for a session after afterID.
func (s *Store) ListEvents(ctx context.Context, sessionID, afterID string) ([]event.Event, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT event_id, session_id, ts, source, kind, attributes FROM events
		 WHERE session_id = ? ORDER BY rowid`, sessionID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []event.Event
	seen := afterID == ""
	for rows.Next() {
		var ev event.Event
		var ts, attrs string
		if err := rows.Scan(&ev.EventID, &ev.SessionID, &ts, &ev.Source, &ev.Kind, &attrs); err != nil {
			return nil, err
		}
		if !seen {
			if ev.EventID == afterID {
				seen = true
			}
			continue
		}
		ev.TS, _ = time.Parse(time.RFC3339Nano, ts)
		ev.Attributes = map[string]any{}
		if err := json.Unmarshal([]byte(attrs), &ev.Attributes); err != nil {
			return nil, err
		}
		out = append(out, ev)
	}
	return out, rows.Err()
}
