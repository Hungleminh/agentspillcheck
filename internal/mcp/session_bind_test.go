// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package mcp

import "testing"

func TestMarketplaceSessionIDUpdatesWhenHeaderArrivesLater(t *testing.T) {
	backends := NewTokenBackends("en", 42)
	if _, id, _ := backends.bind("token-a", ""); id != "" {
		t.Fatal("empty header must not invent a session id")
	}
	backend, id, _ := backends.bind("token-a", "spillcheck-1")
	if id != "spillcheck-1" || backends.MarketplaceSessionID("token-a") != "spillcheck-1" {
		t.Fatalf("session id = %q", id)
	}
	again, _, _ := backends.bind("token-a", "spillcheck-2")
	if again != backend || backends.MarketplaceSessionID("token-a") != "spillcheck-2" {
		t.Fatal("a later verified session id must replace the stale one on the same bearer")
	}
	if _, id, _ = backends.bind("token-a", "  "); backends.MarketplaceSessionID("token-a") != "spillcheck-2" || id != "spillcheck-2" {
		t.Fatal("blank header must keep the last verified session id")
	}
}
