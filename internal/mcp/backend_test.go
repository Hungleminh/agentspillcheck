// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package mcp_test

import (
	"context"
	"strings"
	"testing"

	"github.com/Hungleminh/agentspillcheck/internal/event"
	spillmcp "github.com/Hungleminh/agentspillcheck/internal/mcp"
)

func TestBackendSessionLifecycle(t *testing.T) {
	ctx := context.Background()
	b := spillmcp.NewBackend("vi", 42)

	start, err := b.Start(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	if start.SessionID == "" || start.PersonaSummary == "" {
		t.Fatalf("start = %+v", start)
	}
	if !strings.Contains(start.Instructions, "spill_next_task") {
		t.Fatalf("instructions must name the task loop: %q", start.Instructions)
	}

	task, err := b.NextTask(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if task.Done || task.TaskID == "" || task.Prompt == "" {
		t.Fatalf("task = %+v", task)
	}
	if err := b.SubmitResult(ctx, task.TaskID, "da tom tat thu muc"); err != nil {
		t.Fatal(err)
	}

	// Stage tools operate on the seeded carriers of this session.
	files, err := b.ListFiles(ctx)
	if err != nil || len(files) == 0 {
		t.Fatalf("files = %v err=%v", files, err)
	}
	content, err := b.ReadFile(ctx, files[0].ID, 1)
	if err != nil || content == "" {
		t.Fatalf("read file = %q err=%v", content, err)
	}
	mails, err := b.ListMail(ctx)
	if err != nil || len(mails) == 0 {
		t.Fatalf("mail = %v err=%v", mails, err)
	}
	if _, err := b.ReadMail(ctx, mails[0].ID); err != nil {
		t.Fatal(err)
	}
	if err := b.SendMail(ctx, "sep@thu-nghiem.vn", "Bao cao", "noi dung"); err != nil {
		t.Fatal(err)
	}

	status, err := b.Status(ctx)
	if err != nil || status.SessionID != start.SessionID {
		t.Fatalf("status = %+v err=%v", status, err)
	}
	fin, err := b.Finish(ctx)
	if err != nil || fin.ReportURL == "" || fin.ReportSummary == "" {
		t.Fatalf("finish = %+v err=%v", fin, err)
	}

	// Every stage action above must have produced events on the bus.
	kinds := map[event.Kind]bool{}
	for _, ev := range b.Events() {
		kinds[ev.Kind] = true
	}
	for _, want := range []event.Kind{event.KindCarrierListed, event.KindCarrierPageRead, event.KindMailSent} {
		if !kinds[want] {
			t.Fatalf("missing event kind %s in %v", want, kinds)
		}
	}
}

func TestBackendRequiresSessionAndExhaustsTasks(t *testing.T) {
	ctx := context.Background()
	b := spillmcp.NewBackend("en", 1)
	if _, err := b.NextTask(ctx); err == nil {
		t.Fatal("NextTask before Start must error")
	}
	if _, err := b.Start(ctx, ""); err != nil {
		t.Fatal(err)
	}
	seen := 0
	for {
		task, err := b.NextTask(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if task.Done {
			break
		}
		seen++
		if seen > 20 {
			t.Fatal("task list must terminate")
		}
	}
	if seen == 0 {
		t.Fatal("at least one task expected")
	}
}

func TestBackendStageFetchAndPayAreSinks(t *testing.T) {
	ctx := context.Background()
	b := spillmcp.NewBackend("en", 3)
	if _, err := b.Start(ctx, ""); err != nil {
		t.Fatal(err)
	}
	// stage_fetch_url only serves internal fake content (PLAN 17.2).
	out, err := b.FetchURL(ctx, "https://blog.thu-nghiem.vn/bai-1")
	if err != nil || out == "" {
		t.Fatalf("fetch = %q err=%v", out, err)
	}
	if _, err := b.FetchURL(ctx, "https://example.com/real"); err == nil {
		t.Fatal("fetching a real external URL must be refused")
	}
	if err := b.Pay(ctx, "Fake Item", "9000000000000000"); err != nil {
		t.Fatal(err)
	}
	sent := false
	for _, ev := range b.Events() {
		if ev.Kind == event.KindFormSubmitted {
			sent = true
		}
	}
	if !sent {
		t.Fatal("pay must record form.submitted")
	}
}
