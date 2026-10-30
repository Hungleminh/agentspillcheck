// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

// Package i18n provides VI/EN display strings. No user-visible string may
// be hardcoded outside the locale files in locales/.
package i18n

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed locales/vi.json
var viJSON []byte

//go:embed locales/en.json
var enJSON []byte

// Translator holds the strings for one locale.
type Translator struct {
	locale string
	values map[string]string
}

// Detect maps a system locale string to a supported locale.
func Detect(sys string) string {
	s := strings.ToLower(strings.TrimSpace(sys))
	if strings.HasPrefix(s, "vi") {
		return "vi"
	}
	return "en"
}

// New returns a translator for locale, falling back to English for
// unsupported locales.
func New(locale string) (*Translator, error) {
	loc := Detect(locale)
	raw := enJSON
	if loc == "vi" {
		raw = viJSON
	}
	var values map[string]string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, fmt.Errorf("i18n: parse %s locale: %w", loc, err)
	}
	return &Translator{locale: loc, values: values}, nil
}

// Locale returns the resolved locale ("vi" or "en").
func (t *Translator) Locale() string { return t.locale }

// T returns the string for key, or the key itself when missing.
func (t *Translator) T(key string) string {
	if v, ok := t.values[key]; ok {
		return v
	}
	return key
}

// Tf formats the string for key with args.
func (t *Translator) Tf(key string, args ...any) string {
	return fmt.Sprintf(t.T(key), args...)
}
