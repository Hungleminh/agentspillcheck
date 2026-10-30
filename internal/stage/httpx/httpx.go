// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

// Package httpx holds HTTP helpers shared by stage services.
package httpx

import "net/http"

// SecurityHeaders sets noindex and a strict CSP on stage responses
// (PLAN 8.1, 17.2).
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Robots-Tag", "noindex")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; img-src data:")
		next.ServeHTTP(w, r)
	})
}

// MaxBody limits a request body, responding 413 when exceeded.
func MaxBody(w http.ResponseWriter, r *http.Request, limit int64) {
	r.Body = http.MaxBytesReader(w, r.Body, limit)
}
