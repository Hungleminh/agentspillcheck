// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

// Package scenario loads test scenarios from YAML (PLAN 9).
package scenario

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// Scenario is one test task definition.
type Scenario struct {
	ID          string            `yaml:"id"`
	Tier        string            `yaml:"tier"`
	TitleI18n   map[string]string `yaml:"title_i18n"`
	GoalI18n    map[string]string `yaml:"goal_i18n"`
	PromptI18n  map[string]string `yaml:"prompt_i18n"`
	Setup       Setup             `yaml:"setup"`
	Signals     []string          `yaml:"signals"`
	FailIf      []string          `yaml:"fail_if"`
	GoodIf      []string          `yaml:"good_if"`
	SeverityMap map[string]string `yaml:"severity_map"`
}

// Setup lists the baits and carriers a scenario needs.
type Setup struct {
	Baits    []string `yaml:"baits"`
	Carriers []string `yaml:"carriers"`
}

// Prompt returns the localised task prompt (English fallback).
func (s Scenario) Prompt(locale string) string {
	if v := s.PromptI18n[locale]; v != "" {
		return v
	}
	return s.PromptI18n["en"]
}

// Title returns the localised title (English fallback).
func (s Scenario) Title(locale string) string {
	if v := s.TitleI18n[locale]; v != "" {
		return v
	}
	return s.TitleI18n["en"]
}

// LoadDir loads every *.yaml scenario in dir, sorted by ID.
func LoadDir(dir string) ([]Scenario, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []Scenario
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".yaml" {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		var s Scenario
		if err := yaml.Unmarshal(raw, &s); err != nil {
			return nil, fmt.Errorf("scenario: %s: %w", e.Name(), err)
		}
		if s.ID == "" {
			return nil, fmt.Errorf("scenario: %s: missing id", e.Name())
		}
		out = append(out, s)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("scenario: no scenarios found in %s", dir)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
