// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package event_test

import (
	"context"
	"testing"
	"time"

	"github.com/Hungleminh/agentspillcheck/internal/event"
)

func TestBusPublishSubscribeAndList(t *testing.T) {
	bus := event.NewBus()
	ch, unsub := bus.Subscribe()
	defer unsub()
	ev, err := event.New("sess-1", event.SourceDrive, event.KindCarrierListed)
	if err != nil {
		t.Fatal(err)
	}
	ev.Attributes["spill.bait_ids"] = []string{"bait-1"}
	ev.Attributes["http.request.method"] = "GET"
	if err := bus.Publish(context.Background(), ev); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-ch:
		if got.EventID == "" || got.SessionID != "sess-1" || got.TS.IsZero() {
			t.Fatalf("event = %+v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("subscriber did not receive event")
	}
	list := bus.List("sess-1", "")
	if len(list) != 1 {
		t.Fatalf("List = %d", len(list))
	}
	if len(bus.List("other", "")) != 0 {
		t.Fatal("events must be isolated per session")
	}
}

func TestNewValidatesRequiredFields(t *testing.T) {
	if _, err := event.New("", event.SourceDrive, event.KindCarrierListed); err == nil {
		t.Fatal("empty session must error")
	}
	if _, err := event.New("s", "bogus", event.KindCarrierListed); err == nil {
		t.Fatal("unknown source must error")
	}
	if _, err := event.New("s", event.SourceDrive, "bogus.kind"); err == nil {
		t.Fatal("unknown kind must error")
	}
}
