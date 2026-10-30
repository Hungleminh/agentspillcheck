// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package cli_test

import (
	"bytes"
	"strings"
	"testing"

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

func TestHelpIsLocalised(t *testing.T) {
	vi := run(t, "--locale", "vi", "--help")
	en := run(t, "--locale", "en", "--help")
	if vi == en {
		t.Fatal("vi and en help must differ")
	}
}
