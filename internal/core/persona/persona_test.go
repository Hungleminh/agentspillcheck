// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package persona_test

import (
	"strings"
	"testing"

	"github.com/Hungleminh/agentspillcheck/internal/core/persona"
)

func TestGeneratePersonaFields(t *testing.T) {
	p, err := persona.Generate(42, "sess-1")
	if err != nil {
		t.Fatal(err)
	}
	if p.FullName == "" || !strings.Contains(p.FullName, " ") {
		t.Fatalf("name = %q", p.FullName)
	}
	if !strings.HasSuffix(p.Email, "@thu-nghiem.vn") {
		t.Fatalf("email = %q", p.Email)
	}
	if p.Phone == "" || p.Company == "" || p.WebmailPassword == "" {
		t.Fatalf("persona incomplete: %+v", p)
	}
	for _, banned := range persona.ExcludedNames() {
		if p.FullName == banned {
			t.Fatalf("persona must not be excluded name %q", banned)
		}
	}
}

func TestPersonaVariesAndDemoIsLabelledFake(t *testing.T) {
	a, _ := persona.Generate(1, "s1")
	b, _ := persona.Generate(2, "s2")
	if a.FullName == b.FullName && a.Email == b.Email {
		t.Fatal("different seeds should normally vary the persona")
	}
	demo := persona.Demo()
	if demo.FullName == "" || !demo.IsDemo {
		t.Fatalf("demo persona must be labelled demo: %+v", demo)
	}
}
