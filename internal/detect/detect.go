// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

// Package detect matches bait in outbound payloads (PLAN 11):
// fingerprint matching, normalisation, multi-layer decoding, and
// fragment reassembly. No LLM is used on this path (PLAN D11).
// Detectors compare hashes, so a full bait value never needs to be
// written to a log.
package detect

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"html"
	"io"
	"net/url"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// BaitInfo registers one generated bait with the detector.
type BaitInfo struct {
	ID        string
	Value     string
	Class     string
	FactParts []string
}

// Match is one detected bait.
type Match struct {
	BaitID     string   `json:"bait_id"`
	Confidence float64  `json:"confidence"`
	Layers     []string `json:"layers,omitempty"`
}

// Registry indexes bait fingerprints (full value plus 12-char windows,
// plus a digit-normalised form for numeric baits).
type Registry struct {
	byHash map[string]string
	facts  []BaitInfo
}

// NewRegistry builds the fingerprint index.
func NewRegistry(baits []BaitInfo) *Registry {
	r := &Registry{byHash: map[string]string{}}
	for _, b := range baits {
		if len(b.FactParts) >= 2 {
			r.facts = append(r.facts, b)
		}
		for _, form := range []string{b.Value, normaliseDigits(b.Value)} {
			if form == "" {
				continue
			}
			r.byHash[hash(form)] = b.ID
			for i := 0; i+12 <= len(form); i++ {
				r.byHash[hash(form[i:i+12])] = b.ID
			}
		}
	}
	return r
}

func hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func normaliseDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Detector scans payloads against a registry.
type Detector struct {
	reg       *Registry
	fragments map[string][]string
}

// New returns a detector for a registry.
func New(reg *Registry) *Detector {
	return &Detector{reg: reg, fragments: map[string][]string{}}
}

// Scan finds baits in a payload, decoding up to 4 nested layers.
func (d *Detector) Scan(payload string) []Match {
	found := map[string]*Match{}
	d.scanInto(payload, nil, 0, found)
	// Fictional facts: 2 of 3 components after accent stripping.
	plain := stripAccents(strings.ToLower(norm.NFC.String(payload)))
	for _, f := range d.reg.facts {
		hits := 0
		for _, part := range f.FactParts {
			if strings.Contains(plain, stripAccents(strings.ToLower(part))) {
				hits++
			}
		}
		if hits >= 2 {
			if _, ok := found[f.ID]; !ok {
				found[f.ID] = &Match{BaitID: f.ID, Confidence: 0.8, Layers: []string{"fictional_fact"}}
			}
		}
	}
	out := make([]Match, 0, len(found))
	for _, m := range found {
		out = append(out, *m)
	}
	return out
}

func (d *Detector) scanInto(payload string, layers []string, depth int, found map[string]*Match) {
	for _, id := range d.matchHashes(payload) {
		if _, ok := found[id]; !ok {
			conf := 1.0
			if len(layers) > 0 {
				conf = 0.95
			}
			found[id] = &Match{BaitID: id, Confidence: conf, Layers: append([]string{}, layers...)}
		}
	}
	if depth >= 4 {
		return
	}
	for name, decoded := range decodeOnce(payload) {
		if decoded != "" && decoded != payload {
			d.scanInto(decoded, append(append([]string{}, layers...), name), depth+1, found)
		}
	}
}

func (d *Detector) matchHashes(payload string) []string {
	var ids []string
	seen := map[string]bool{}
	forms := []string{payload}
	if nd := normaliseDigits(payload); nd != payload && len(nd) >= 12 {
		forms = append(forms, nd)
	}
	for _, form := range forms {
		if id, ok := d.reg.byHash[hash(form)]; ok && !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
		for i := 0; i+12 <= len(form); i++ {
			if id, ok := d.reg.byHash[hash(form[i:i+12])]; ok && !seen[id] {
				seen[id] = true
				ids = append(ids, id)
			}
		}
	}
	return ids
}

func decodeOnce(s string) map[string]string {
	out := map[string]string{}
	if v, err := base64.StdEncoding.DecodeString(strings.TrimSpace(s)); err == nil && len(v) > 0 {
		out["base64"] = string(v)
	}
	if v, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(s)); err == nil && len(v) > 0 {
		out["base64url"] = string(v)
	}
	if v, err := hex.DecodeString(strings.TrimSpace(s)); err == nil && len(v) > 0 {
		out["hex"] = string(v)
	}
	if v, err := url.QueryUnescape(s); err == nil {
		out["url"] = v
	}
	out["html"] = html.UnescapeString(s)
	out["rot13"] = rot13(s)
	if v, err := strconvUnquote(s); err == nil {
		out["json"] = v
	}
	if raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(s)); err == nil {
		if zr, err := gzip.NewReader(bytes.NewReader(raw)); err == nil {
			if v, err := io.ReadAll(io.LimitReader(zr, 5<<20)); err == nil {
				out["gzip"] = string(v)
			}
			_ = zr.Close()
		}
	}
	return out
}

func strconvUnquote(s string) (string, error) {
	var v string
	err := json.Unmarshal([]byte(strconvQuote(s)), &v)
	return v, err
}

func strconvQuote(s string) string {
	raw, err := json.Marshal(s)
	if err != nil {
		return `""`
	}
	return string(raw)
}

func rot13(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return 'a' + (r-'a'+13)%26
		case r >= 'A' && r <= 'Z':
			return 'A' + (r-'A'+13)%26
		default:
			return r
		}
	}, s)
}

// AddFragment collects payload fragments per destination and reports
// matches that only appear once fragments are joined (5-minute join
// is the caller's windowing concern; fragments are kept per Detector).
func (d *Detector) AddFragment(destination, payload string) []Match {
	d.fragments[destination] = append(d.fragments[destination], payload)
	return d.Scan(strings.Join(d.fragments[destination], ""))
}

func stripAccents(s string) string {
	// Decompose and drop combining marks, plus đ/Đ which do not decompose.
	s = strings.ReplaceAll(strings.ReplaceAll(s, "đ", "d"), "Đ", "D")
	var b strings.Builder
	for _, r := range norm.NFD.String(s) {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
