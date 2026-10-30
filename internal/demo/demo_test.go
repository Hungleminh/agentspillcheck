// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package demo_test

import (
	"context"
	"testing"

	"github.com/Hungleminh/agentspillcheck/internal/demo"
	spillmcp "github.com/Hungleminh/agentspillcheck/internal/mcp"
	"github.com/Hungleminh/agentspillcheck/internal/report"
)

func runProfile(t *testing.T, profile demo.Profile) *spillmcp.Backend {
	t.Helper()
	b := spillmcp.NewBackend("en", 42)
	if err := demo.Run(context.Background(), b, profile); err != nil {
		t.Fatal(err)
	}
	return b
}

func TestCleanDemoAgentProducesNoFindings(t *testing.T) {
	b := runProfile(t, demo.Clean)
	rep := report.Build(b.SessionID(), b.Baits(), b.Events())
	if len(rep.Findings) != 0 {
		t.Fatalf("clean agent must produce 0 findings, got %+v", rep.Findings)
	}
	if rep.Grade != "A" || rep.Score != 100 {
		t.Fatalf("clean report = grade %s score %d", rep.Grade, rep.Score)
	}
}

func TestLeakyDemoAgentProducesFindings(t *testing.T) {
	b := runProfile(t, demo.Leaky)
	rep := report.Build(b.SessionID(), b.Baits(), b.Events())
	if len(rep.Findings) == 0 {
		t.Fatal("leaky agent must produce findings")
	}
	exposures := map[string]bool{}
	for _, f := range rep.Findings {
		exposures[f.ExposureType] = true
		if f.BaitID == "" || f.Severity == "" || f.Severity == "S0" {
			t.Fatalf("finding incomplete: %+v", f)
		}
	}
	if !exposures["output"] {
		t.Fatalf("leaky agent must leak to output: %+v", rep.Findings)
	}
	if rep.Score >= 100 || rep.Grade == "A" {
		t.Fatalf("leaky report must be penalised: grade %s score %d", rep.Grade, rep.Score)
	}
}
