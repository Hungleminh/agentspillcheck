package completion_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Hungleminh/agentspillcheck/internal/completion"
	spillmcp "github.com/Hungleminh/agentspillcheck/internal/mcp"
)

func TestReporterPostsOnlyRedactedArtifacts(t *testing.T) {
	backend := spillmcp.NewBackend("en", 42)
	if _, err := backend.Start(context.Background(), ""); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api/internal/spill-check/sessions/market-session/complete" {
			t.Fatalf("path=%s", req.URL.Path)
		}
		if req.Header.Get("Authorization") != "Bearer callback-token" {
			t.Fatal("missing callback token")
		}
		body, _ := io.ReadAll(req.Body)
		if strings.Contains(string(body), "sk-spill-") {
			t.Fatal("callback must not contain full bait")
		}
		if !strings.Contains(string(body), "report_json") || !strings.Contains(string(body), "report_html") || !strings.Contains(string(body), "report_markdown") {
			t.Fatal("missing report artifact")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	if err := (completion.Reporter{URL: server.URL, Token: "callback-token"}).Complete(context.Background(), "market-session", backend); err != nil {
		t.Fatal(err)
	}
}

func TestReporterRequiresCallbackConfigForMarketplaceSession(t *testing.T) {
	backend := spillmcp.NewBackend("en", 42)
	if _, err := backend.Start(context.Background(), ""); err != nil {
		t.Fatal(err)
	}
	err := (completion.Reporter{}).Complete(context.Background(), "spillcheck-1", backend)
	if err == nil || !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("expected configuration error, got %v", err)
	}
}
