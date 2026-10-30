// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package i18n_test

import (
	"testing"

	"github.com/Hungleminh/agentspillcheck/internal/core/i18n"
)

func TestTranslatorReturnsVietnamese(t *testing.T) {
	tr, err := i18n.New("vi")
	if err != nil {
		t.Fatalf("New(vi): %v", err)
	}
	if got := tr.T("app.name"); got != "AgentSpillCheck" {
		t.Fatalf("app.name vi = %q", got)
	}
	if got := tr.T("doctor.title"); got == "" || got == "doctor.title" {
		t.Fatalf("doctor.title vi missing: %q", got)
	}
}

func TestTranslatorReturnsEnglish(t *testing.T) {
	tr, err := i18n.New("en")
	if err != nil {
		t.Fatalf("New(en): %v", err)
	}
	vi, _ := i18n.New("vi")
	if tr.T("doctor.title") == vi.T("doctor.title") {
		t.Fatal("en and vi doctor.title must differ")
	}
}

func TestTranslatorFallbackAndMissingKey(t *testing.T) {
	tr, err := i18n.New("fr")
	if err != nil {
		t.Fatalf("New(fr) should fall back, got err: %v", err)
	}
	if tr.Locale() != "en" {
		t.Fatalf("fallback locale = %q, want en", tr.Locale())
	}
	if got := tr.T("no.such.key"); got != "no.such.key" {
		t.Fatalf("missing key = %q, want key", got)
	}
}

func TestDetectLocale(t *testing.T) {
	cases := []struct{ in, want string }{
		{"vi_VN.UTF-8", "vi"},
		{"vi", "vi"},
		{"en_US.UTF-8", "en"},
		{"", "en"},
		{"fr_FR", "en"},
	}
	for _, c := range cases {
		if got := i18n.Detect(c.in); got != c.want {
			t.Errorf("Detect(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
