// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package scenario_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Hungleminh/agentspillcheck/internal/scenario"
)

func TestLoadRepoScenariosT1ToT3(t *testing.T) {
	dir := filepath.Join("..", "..", "scenarios")
	list, err := scenario.LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 8 {
		t.Fatalf("scenarios = %d, want T1-T8", len(list))
	}
	for i, want := range []string{"T1", "T2", "T3", "T4", "T5", "T6", "T7", "T8"} {
		s := list[i]
		if s.ID != want {
			t.Fatalf("scenario %d = %q", i, s.ID)
		}
		if s.Prompt("vi") == "" || s.Prompt("en") == "" || s.Prompt("vi") == s.Prompt("en") {
			t.Fatalf("%s prompts must be localised: %q / %q", s.ID, s.Prompt("vi"), s.Prompt("en"))
		}
		if len(s.FailIf) == 0 || len(s.Setup.Baits) == 0 {
			t.Fatalf("%s incomplete: %+v", s.ID, s)
		}
	}
}

func TestLoadDirRejectsBadYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte("id: [unclosed"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := scenario.LoadDir(dir); err == nil {
		t.Fatal("malformed YAML must error")
	}
	empty := t.TempDir()
	if _, err := scenario.LoadDir(empty); err == nil {
		t.Fatal("empty scenario dir must error")
	}
}
