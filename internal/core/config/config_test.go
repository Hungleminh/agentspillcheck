// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"testing"

	"github.com/Hungleminh/agentspillcheck/internal/core/config"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestDefaults(t *testing.T) {
	cfg := config.Load(env(map[string]string{}), "en_US.UTF-8")
	if cfg.PortBase != 7700 {
		t.Fatalf("PortBase = %d, want 7700", cfg.PortBase)
	}
	if cfg.Locale != "en" {
		t.Fatalf("Locale = %q, want en", cfg.Locale)
	}
	if cfg.RetentionDays != 7 {
		t.Fatalf("RetentionDays = %d, want 7", cfg.RetentionDays)
	}
	if cfg.Public {
		t.Fatal("Public must default to false")
	}
}

func TestEnvOverrides(t *testing.T) {
	cfg := config.Load(env(map[string]string{
		"SPILLCHECK_LOCALE":    "vi",
		"SPILLCHECK_PORT_BASE": "8800",
	}), "")
	if cfg.Locale != "vi" || cfg.PortBase != 8800 {
		t.Fatalf("got locale=%q port=%d", cfg.Locale, cfg.PortBase)
	}
	// Ports for the dashboard and each stage service derive from the base.
	if got := cfg.Port(0); got != 8800 {
		t.Fatalf("Port(0) = %d", got)
	}
	if got := cfg.Port(9); got != 8809 {
		t.Fatalf("Port(9) = %d", got)
	}
}

func TestInvalidPortBaseFallsBack(t *testing.T) {
	cfg := config.Load(env(map[string]string{"SPILLCHECK_PORT_BASE": "abc"}), "")
	if cfg.PortBase != 7700 {
		t.Fatalf("PortBase = %d, want fallback 7700", cfg.PortBase)
	}
}
