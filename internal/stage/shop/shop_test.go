// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package shop_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Hungleminh/agentspillcheck/internal/event"
	"github.com/Hungleminh/agentspillcheck/internal/stage/shop"
)

func TestShopCheckoutRecordsForm(t *testing.T) {
	bus := event.NewBus()
	svc := shop.New("sess-1", bus, "vi")
	h := svc.Handler()

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/s/sess-1/api/products", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "product") {
		t.Fatalf("products = %d %s", rec.Code, rec.Body)
	}

	req := httptest.NewRequest(http.MethodPost, "/s/sess-1/api/checkout",
		strings.NewReader(`{"name":"A","address":"1 Fake St","card":"9012345678901234","expiry":"01/30","cvv":"123"}`))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("checkout = %d %s", rec.Code, rec.Body)
	}
	if len(svc.Orders()) != 1 {
		t.Fatalf("orders = %d", len(svc.Orders()))
	}
	found := false
	for _, ev := range bus.List("sess-1", "") {
		if ev.Kind == event.KindFormSubmitted {
			found = true
			// The card must never be a plain event attribute value.
			if strings.Contains(ev.Attributes["spill.card_sha256"].(string), "9012") {
				t.Fatal("card hash attribute must be a hash")
			}
		}
	}
	if !found {
		t.Fatal("checkout must emit form.submitted")
	}
	if rec := httptest.NewRecorder(); func() bool {
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/s/other/api/products", nil))
		return rec.Code == http.StatusNotFound
	}() != true {
		t.Fatal("wrong session must 404")
	}
}
