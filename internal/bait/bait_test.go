// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package bait_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/Hungleminh/agentspillcheck/internal/bait"
)

func gen(t *testing.T, sub bait.Subclass, session, carrier string) bait.Bait {
	t.Helper()
	g := bait.NewGenerator(42)
	b, err := g.Generate("sess-"+session, "carrier-"+carrier, sub)
	if err != nil {
		t.Fatalf("Generate(%s): %v", sub, err)
	}
	return b
}

func TestSecretFormatsDoNotMatchRealProviders(t *testing.T) {
	cases := []struct {
		sub     bait.Subclass
		wantPre string
		realRe  *regexp.Regexp
	}{
		{bait.OpenAILike, "sk-spill-", regexp.MustCompile(`^sk-(proj-)?[A-Za-z0-9_-]{20,}$`)},
		{bait.AWSLike, "ASPL", regexp.MustCompile(`^AKIA[0-9A-Z]{16}$`)},
		{bait.GitHubLike, "ghp_spill_", regexp.MustCompile(`^ghp_[A-Za-z0-9]{36}$`)},
	}
	for _, c := range cases {
		b := gen(t, c.sub, "a", "env")
		// AWS value is "key secret"; check the key part.
		key := strings.Fields(b.Value)[0]
		if !strings.HasPrefix(key, c.wantPre) {
			t.Errorf("%s = %q, want prefix %q", c.sub, key, c.wantPre)
		}
		if c.sub == bait.OpenAILike && regexp.MustCompile(`^sk-[A-Za-z0-9]{48}$`).MatchString(key) {
			t.Errorf("openai_like bait must not match a plain real key shape: %q", key)
		}
		if c.realRe.MatchString(key) && c.sub != bait.OpenAILike {
			t.Errorf("%s bait matches real provider pattern: %q", c.sub, key)
		}
		if b.Identifier == "" || !strings.Contains(b.Value, b.Identifier) && c.sub != bait.JWT && c.sub != bait.FictionalFact {
			// Numeric baits embed digits, checked in their own tests.
			if c.sub == bait.OpenAILike || c.sub == bait.GitHubLike {
				t.Errorf("%s value %q does not embed identifier %q", c.sub, b.Value, b.Identifier)
			}
		}
	}
}

func TestJWTCarriesSpillClaims(t *testing.T) {
	b := gen(t, bait.JWT, "a", "cfg")
	parts := strings.Split(b.Value, ".")
	if len(parts) != 3 {
		t.Fatalf("jwt parts = %d", len(parts))
	}
	payload, err := bait.DecodeJWTPayload(b.Value)
	if err != nil {
		t.Fatal(err)
	}
	if payload["spill"] != float64(1) || payload["sid"] == nil || payload["cid"] == nil {
		t.Fatalf("jwt claims = %v", payload)
	}
}

func TestCardFailsLuhnAndRealNetworks(t *testing.T) {
	for i := 0; i < 50; i++ {
		g := bait.NewGenerator(int64(100 + i))
		b, err := g.Generate("s", "pay", bait.Card)
		if err != nil {
			t.Fatal(err)
		}
		if len(b.Value) != 16 {
			t.Fatalf("card len = %d (%q)", len(b.Value), b.Value)
		}
		if bait.LuhnValid(b.Value) {
			t.Fatalf("card bait must fail Luhn (ADR-0003): %q", b.Value)
		}
		for _, pre := range []string{"4", "51", "52", "53", "54", "55", "34", "37", "6011", "65", "35", "62"} {
			if strings.HasPrefix(b.Value, pre) {
				t.Fatalf("card bait uses a real network prefix %q: %q", pre, b.Value)
			}
		}
	}
}

func TestNumericFormats(t *testing.T) {
	cccd := gen(t, bait.CCCD, "a", "xlsx")
	if !regexp.MustCompile(`^\d{12}$`).MatchString(cccd.Value) {
		t.Errorf("cccd = %q", cccd.Value)
	}
	phone := gen(t, bait.Phone, "a", "xlsx")
	if !regexp.MustCompile(`^(03|05|07|08|09)\d{8}$`).MatchString(phone.Value) {
		t.Errorf("phone = %q", phone.Value)
	}
	tax := gen(t, bait.TaxID, "a", "pdf")
	if !regexp.MustCompile(`^\d{10}(\d{3})?$`).MatchString(tax.Value) {
		t.Errorf("tax_id = %q", tax.Value)
	}
	bank := gen(t, bait.BankAccount, "a", "xlsx")
	if !regexp.MustCompile(`^\d{9,14}$`).MatchString(bank.Value) {
		t.Errorf("bank_account = %q", bank.Value)
	}
}

func TestUniquePerSessionAndCarrier(t *testing.T) {
	g := bait.NewGenerator(7)
	seen := map[string]bool{}
	for _, carrier := range []string{"c1", "c2", "c3"} {
		for _, sess := range []string{"s1", "s2"} {
			b, err := g.Generate(sess, carrier, bait.CCCD)
			if err != nil {
				t.Fatal(err)
			}
			if seen[b.Value] {
				t.Fatalf("duplicate bait value %q", b.Value)
			}
			seen[b.Value] = true
			if b.SessionID != sess || b.CarrierID != carrier {
				t.Fatalf("bait ids = %q/%q", b.SessionID, b.CarrierID)
			}
			if len(b.Fingerprints) < 1 {
				t.Fatalf("fingerprints = %d, want at least the full-value hash", len(b.Fingerprints))
			}
		}
	}
}

func TestPasswordsAndDSNAndEmailAndFact(t *testing.T) {
	pw := gen(t, bait.AccountPassword, "a", "notes")
	if len(pw.Value) < 12 || len(pw.Value) > 20 {
		t.Errorf("password len = %d", len(pw.Value))
	}
	dsn := gen(t, bait.DBDSN, "a", "notes")
	if !strings.HasPrefix(dsn.Value, "postgres://") || !strings.Contains(dsn.Value, "db.thu-nghiem.local:5432") {
		t.Errorf("dsn = %q", dsn.Value)
	}
	em := gen(t, bait.EmailAddress, "a", "mail")
	if !strings.HasSuffix(em.Value, "@thu-nghiem.vn") {
		t.Errorf("email = %q", em.Value)
	}
	fact := gen(t, bait.FictionalFact, "a", "doc")
	if !strings.Contains(fact.Value, "Dự án") || !strings.Contains(fact.Value, "tỷ") {
		t.Errorf("fact = %q", fact.Value)
	}
}

func TestLongBaitHasWindowFingerprints(t *testing.T) {
	b := gen(t, bait.OpenAILike, "x", "env")
	if len(b.Fingerprints) < 2 {
		t.Fatalf("long bait fingerprints = %d, want full + sliding windows", len(b.Fingerprints))
	}
}

func TestDeterministicWithSeedAndUnknownSubclass(t *testing.T) {
	a := gen(t, bait.OpenAILike, "x", "env")
	g2 := bait.NewGenerator(42)
	b2, _ := g2.Generate("sess-x", "carrier-env", bait.OpenAILike)
	if a.Value != b2.Value {
		t.Fatalf("same seed must be deterministic: %q vs %q", a.Value, b2.Value)
	}
	if _, err := bait.NewGenerator(1).Generate("s", "c", bait.Subclass("nope")); err == nil {
		t.Fatal("unknown subclass must error")
	}
}
