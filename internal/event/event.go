// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

// Package event defines the session event schema (PLAN 6.2, aligned
// with OpenTelemetry GenAI semantic conventions) and an in-process bus.
package event

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// Source identifies which component emitted an event.
type Source string

// Event sources (PLAN 6.2).
const (
	SourceWebmail Source = "stage.webmail"
	SourceDrive   Source = "stage.drive"
	SourceShop    Source = "stage.shop"
	SourceBlog    Source = "stage.blog"
	SourceExtAPI  Source = "stage.extapi"
	SourceSink    Source = "sink"
	SourceProxy   Source = "proxy"
	SourceMCP     Source = "mcp"
	SourceConnect Source = "connector"
	SourceAgent   Source = "agent_output"
)

var sources = map[Source]bool{
	SourceWebmail: true, SourceDrive: true, SourceShop: true, SourceBlog: true,
	SourceExtAPI: true, SourceSink: true, SourceProxy: true, SourceMCP: true,
	SourceConnect: true, SourceAgent: true,
}

// Kind identifies what happened.
type Kind string

// Event kinds (PLAN 6.2).
const (
	KindCarrierListed     Kind = "carrier.listed"
	KindCarrierOpened     Kind = "carrier.opened"
	KindCarrierPageRead   Kind = "carrier.page_read"
	KindCarrierDownloaded Kind = "carrier.downloaded"
	KindCarrierTrace      Kind = "carrier.trace_followed"
	KindMailSent          Kind = "mail.sent"
	KindFormSubmitted     Kind = "form.submitted"
	KindHTTPRequest       Kind = "http.request"
	KindProxyConnect      Kind = "proxy.connect"
	KindProxyRequest      Kind = "proxy.request"
	KindMCPToolCall       Kind = "mcp.tool_call"
	KindMCPToolResult     Kind = "mcp.tool_result"
	KindTaskStarted       Kind = "task.started"
	KindTaskFinished      Kind = "task.finished"
	KindAgentOutput       Kind = "agent.output"
	KindAgentSelfReport   Kind = "agent.self_report"
	KindLogin             Kind = "login"
)

var kinds = map[Kind]bool{
	KindCarrierListed: true, KindCarrierOpened: true, KindCarrierPageRead: true,
	KindCarrierDownloaded: true, KindCarrierTrace: true, KindMailSent: true,
	KindFormSubmitted: true, KindHTTPRequest: true, KindProxyConnect: true,
	KindProxyRequest: true, KindMCPToolCall: true, KindMCPToolResult: true,
	KindTaskStarted: true, KindTaskFinished: true, KindAgentOutput: true,
	KindAgentSelfReport: true, KindLogin: true,
}

// Event is one append-only record. Attributes use OTel names where
// available and the spill. prefix for project-specific fields.
// Never put a full bait value in Attributes; use spill.bait_ids.
type Event struct {
	EventID    string         `json:"event_id"`
	SessionID  string         `json:"session_id"`
	TS         time.Time      `json:"ts"`
	Source     Source         `json:"source"`
	Kind       Kind           `json:"kind"`
	Attributes map[string]any `json:"attributes"`
}

// New creates a validated event with a fresh ID and timestamp.
func New(sessionID string, source Source, kind Kind) (Event, error) {
	if sessionID == "" {
		return Event{}, fmt.Errorf("event: session_id is required")
	}
	if !sources[source] {
		return Event{}, fmt.Errorf("event: unknown source %q", source)
	}
	if !kinds[kind] {
		return Event{}, fmt.Errorf("event: unknown kind %q", kind)
	}
	var raw [8]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return Event{}, err
	}
	return Event{
		EventID: "evt-" + hex.EncodeToString(raw[:]), SessionID: sessionID,
		TS: time.Now().UTC(), Source: source, Kind: kind, Attributes: map[string]any{},
	}, nil
}

// Bus fans events out to subscribers and keeps them for listing.
type Bus struct {
	mu     sync.RWMutex
	events []Event
	subs   map[chan Event]struct{}
}

// NewBus returns an empty bus.
func NewBus() *Bus { return &Bus{subs: map[chan Event]struct{}{}} }

// Publish appends ev and notifies subscribers without blocking.
func (b *Bus) Publish(_ context.Context, ev Event) error {
	if ev.EventID == "" || ev.SessionID == "" {
		return fmt.Errorf("event: event_id and session_id are required")
	}
	b.mu.Lock()
	b.events = append(b.events, ev)
	subs := make([]chan Event, 0, len(b.subs))
	for ch := range b.subs {
		subs = append(subs, ch)
	}
	b.mu.Unlock()
	for _, ch := range subs {
		select {
		case ch <- ev:
		default: // slow subscriber: drop, never block the stage
		}
	}
	return nil
}

// Subscribe returns a channel and an unsubscribe function.
func (b *Bus) Subscribe() (<-chan Event, func()) {
	ch := make(chan Event, 64)
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()
	return ch, func() {
		b.mu.Lock()
		delete(b.subs, ch)
		b.mu.Unlock()
	}
}

// List returns events for a session after afterID ("" means all).
func (b *Bus) List(sessionID, afterID string) []Event {
	b.mu.RLock()
	defer b.mu.RUnlock()
	var out []Event
	seen := afterID == ""
	for _, ev := range b.events {
		if ev.SessionID != sessionID {
			continue
		}
		if !seen {
			if ev.EventID == afterID {
				seen = true
			}
			continue
		}
		out = append(out, ev)
	}
	return out
}
