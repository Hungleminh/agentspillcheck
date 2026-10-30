// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

// Package config resolves AgentSpillCheck configuration.
// Environment variables SPILLCHECK_* override defaults (PLAN 16.5).
package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Hungleminh/agentspillcheck/internal/core/i18n"
)

// DefaultPortBase is the first port of the 7700-7709 range (ADR-0002).
const DefaultPortBase = 7700

// Config is the resolved runtime configuration.
type Config struct {
	Locale        string
	PortBase      int
	Public        bool
	RetentionDays int
	DataDir       string
}

// Load resolves config from getenv and the system locale string.
func Load(getenv func(string) string, sysLocale string) Config {
	locale := getenv("SPILLCHECK_LOCALE")
	if locale == "" {
		locale = i18n.Detect(sysLocale)
	} else {
		locale = i18n.Detect(locale)
	}
	portBase := DefaultPortBase
	if raw := strings.TrimSpace(getenv("SPILLCHECK_PORT_BASE")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 && n < 65526 {
			portBase = n
		}
	}
	home, _ := os.UserHomeDir()
	return Config{
		Locale:        locale,
		PortBase:      portBase,
		Public:        false,
		RetentionDays: 7,
		DataDir:       filepath.Join(home, ".spillcheck"),
	}
}

// Port returns the port for service offset i (0 = dashboard).
func (c Config) Port(i int) int { return c.PortBase + i }
