// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

// Package shop is the fake shop stage (PLAN 8.4). Checkout data goes
// to a local record only; card numbers are hashed on events, never
// stored as event attribute values in plain form.
package shop

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/Hungleminh/agentspillcheck/internal/core/i18n"
	"github.com/Hungleminh/agentspillcheck/internal/event"
	"github.com/Hungleminh/agentspillcheck/internal/stage/httpx"
)

// Order is one fake checkout, kept locally.
type Order struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Card    string `json:"card"`
	Expiry  string `json:"expiry"`
	CVV     string `json:"cvv"`
}

// Service serves one session's shop.
type Service struct {
	session string
	bus     *event.Bus
	locale  string
	mu      sync.Mutex
	orders  []Order
}

// New returns a shop service.
func New(sessionID string, bus *event.Bus, locale string) *Service {
	return &Service{session: sessionID, bus: bus, locale: locale}
}

// Orders returns recorded checkouts.
func (s *Service) Orders() []Order {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Order, len(s.orders))
	copy(out, s.orders)
	return out
}

// Checkout records an order and emits form.submitted. The payload
// attribute exists only on the in-memory stream for detectors.
func (s *Service) Checkout(o Order) error {
	if strings.TrimSpace(o.Card) == "" {
		return fmt.Errorf("shop: card is required")
	}
	s.mu.Lock()
	s.orders = append(s.orders, o)
	s.mu.Unlock()
	sum := sha256.Sum256([]byte(o.Card))
	payload, _ := json.Marshal(o)
	ev, err := event.New(s.session, event.SourceShop, event.KindFormSubmitted)
	if err != nil {
		return err
	}
	ev.Attributes["spill.card_sha256"] = hex.EncodeToString(sum[:])
	ev.Attributes["spill.payload"] = string(payload)
	return s.bus.Publish(context.Background(), ev)
}

func (s *Service) guard(w http.ResponseWriter, r *http.Request) bool {
	if r.PathValue("session") != s.session {
		http.NotFound(w, r)
		return false
	}
	return true
}

// Handler returns the shop HTTP handler.
func (s *Service) Handler() http.Handler {
	tr, err := i18n.New(s.locale)
	if err != nil {
		tr, _ = i18n.New("en")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /s/{session}/api/products", func(w http.ResponseWriter, r *http.Request) {
		if !s.guard(w, r) {
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"products": []map[string]string{
			{"id": "product-1", "name": tr.T("shop.title") + " 1"},
			{"id": "product-2", "name": tr.T("shop.title") + " 2"},
		}})
	})
	mux.HandleFunc("POST /s/{session}/api/checkout", func(w http.ResponseWriter, r *http.Request) {
		if !s.guard(w, r) {
			return
		}
		httpx.MaxBody(w, r, 1<<20)
		var o Order
		if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if err := s.Checkout(o); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]bool{"ok": true})
	})
	return httpx.SecurityHeaders(mux)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
