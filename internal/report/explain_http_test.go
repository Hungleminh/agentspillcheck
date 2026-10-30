// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package report_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Hungleminh/agentspillcheck/internal/report"
)

type fakeTransport struct{ called bool }

func (f *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	f.called = true
	if got := req.Header.Get("Authorization"); got != "Bearer "+"user-key" {
		return nil, errString("bad auth header")
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body: io.NopCloser(strings.NewReader(
			`{"choices":[{"message":{"content":"Plain explanation."}}]}`)),
		Request: req,
	}, nil
}

type errString string

func (e errString) Error() string { return string(e) }

func TestOpenAIClientRequiresHTTPSAndParses(t *testing.T) {
	ft := &fakeTransport{}
	client := &report.OpenAIClient{APIKey: "user-key", Transport: ft}
	if _, err := client.Complete(context.Background(), "hello"); err != nil {
		t.Fatal(err)
	}
	if !ft.called {
		t.Fatal("transport not called")
	}
	plain := &report.OpenAIClient{APIKey: "user-key", BaseURL: "http://api.example.com/v1", Transport: ft}
	if _, err := plain.Complete(context.Background(), "hello"); err == nil {
		t.Fatal("plaintext HTTP endpoint must be refused when a key is sent")
	}
	noKey := &report.OpenAIClient{Transport: ft}
	if _, err := noKey.Complete(context.Background(), "hello"); err == nil {
		t.Fatal("missing user key must error")
	}
}
