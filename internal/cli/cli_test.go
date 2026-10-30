// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package cli_test

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Hungleminh/agentspillcheck/internal/cli"
)

func run(t *testing.T, args ...string) string {
	t.Helper()
	var out bytes.Buffer
	cmd := cli.NewRootCommand("v0.0.0-test", &out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute(%v): %v (out=%q)", args, err, out.String())
	}
	return out.String()
}

func TestVersionCommand(t *testing.T) {
	got := run(t, "version")
	if !strings.Contains(got, "v0.0.0-test") {
		t.Fatalf("version output = %q", got)
	}
}

func TestDoctorCommandReportsChecks(t *testing.T) {
	got := run(t, "doctor", "--locale", "en")
	for _, want := range []string{"Go", "7700", "data"} {
		if !strings.Contains(got, want) {
			t.Fatalf("doctor output missing %q: %q", want, got)
		}
	}
}

func TestRootHasRequiredSubcommands(t *testing.T) {
	cmd := cli.NewRootCommand("v0.0.0-test", &bytes.Buffer{})
	want := map[string]bool{"version": false, "doctor": false, "up": false}
	for _, c := range cmd.Commands() {
		if _, ok := want[c.Name()]; ok {
			want[c.Name()] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("missing subcommand %q", name)
		}
	}
}

func TestDemoAndBaitPreviewCommands(t *testing.T) {
	got := run(t, "demo", "--locale", "en")
	for _, want := range []string{"grade", "score"} {
		if !strings.Contains(strings.ToLower(got), want) {
			t.Fatalf("demo output missing %q: %q", want, got)
		}
	}
	got = run(t, "bait", "preview", "--tier", "standard")
	if !strings.Contains(got, "openai_like") || !strings.Contains(got, "cccd") {
		t.Fatalf("bait preview = %q", got)
	}
	got = run(t, "session", "new", "--locale", "en", "--tier", "standard")
	if !strings.Contains(got, "sess-") {
		t.Fatalf("session new = %q", got)
	}
}

func TestPurgeAndDownCommands(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	old := home + "/.spillcheck/sessions/old-session"
	if err := os.MkdirAll(old, 0o700); err != nil {
		t.Fatal(err)
	}
	f := old + "/report.json"
	if err := os.WriteFile(f, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	past := time.Now().Add(-10 * 24 * time.Hour)
	if err := os.Chtimes(f, past, past); err != nil {
		t.Fatal(err)
	}
	got := run(t, "purge", "--older-than", "24h")
	if _, err := os.Stat(f); !os.IsNotExist(err) {
		t.Fatalf("purge must remove old session artefacts, stat err=%v out=%q", err, got)
	}
	got = run(t, "down")
	if got == "" {
		t.Fatal("down with no running instance must still print a localised message")
	}
}

func TestMCPConfigCommand(t *testing.T) {
	got := run(t, "mcp", "config", "--client", "claude-desktop")
	for _, want := range []string{"spillcheck", "mcpServers", "mcp", "serve"} {
		if !strings.Contains(got, want) {
			t.Fatalf("mcp config missing %q: %q", want, got)
		}
	}
}

func TestHelpIsLocalised(t *testing.T) {
	vi := run(t, "--locale", "vi", "--help")
	en := run(t, "--locale", "en", "--help")
	if vi == en {
		t.Fatal("vi and en help must differ")
	}
}
