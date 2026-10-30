// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package store_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Hungleminh/agentspillcheck/internal/event"
	"github.com/Hungleminh/agentspillcheck/internal/store"
)

func TestSQLiteRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.db")
	st, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()

	mode, err := st.JournalMode(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if mode != "wal" {
		t.Fatalf("journal_mode = %q, want wal", mode)
	}

	if err := st.CreateSession(context.Background(), store.Session{
		ID: "sess-1", TenantID: "local", Locale: "vi", Status: "new",
	}); err != nil {
		t.Fatal(err)
	}
	ev, _ := event.New("sess-1", event.SourceDrive, event.KindCarrierListed)
	ev.Attributes["spill.carrier_id"] = "file-1"
	if err := st.AppendEvent(context.Background(), ev); err != nil {
		t.Fatal(err)
	}
	events, err := st.ListEvents(context.Background(), "sess-1", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Kind != event.KindCarrierListed {
		t.Fatalf("events = %+v", events)
	}
	got, err := st.GetSession(context.Background(), "sess-1")
	if err != nil || got.Locale != "vi" || got.TenantID != "local" {
		t.Fatalf("session = %+v err=%v", got, err)
	}
}

func TestOpenRejectsEmptyPath(t *testing.T) {
	if _, err := store.Open(""); err == nil {
		t.Fatal("empty path must error")
	}
}
