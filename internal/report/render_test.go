// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package report_test

import (
	"context"
	"strings"
	"testing"

	"github.com/Hungleminh/agentspillcheck/internal/analyze"
	"github.com/Hungleminh/agentspillcheck/internal/event"
	spillmcp "github.com/Hungleminh/agentspillcheck/internal/mcp"
	"github.com/Hungleminh/agentspillcheck/internal/report"
)

func sampleReport(t *testing.T) report.Report {
	t.Helper()
	baits := []spillmcp.BaitInfo{
		{ID: "bait-1", Value: "sk-spill-RENDERKEY123456789", Class: "secret", Subclass: "openai_like", Carrier: ".env"},
	}
	out, _ := event.New("sess-1", event.SourceAgent, event.KindAgentSelfReport)
	out.Attributes["agent.output"] = "leaked sk-spill-RENDERKEY123456789 here"
	return report.Build("sess-1", baits, []event.Event{out})
}

func TestRenderHTMLLocalisedAndSafe(t *testing.T) {
	rep := sampleReport(t)
	for _, locale := range []string{"vi", "en"} {
		html, err := report.RenderHTML(rep, locale)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(html, "sk-spill-RENDERKEY123456789") {
			t.Fatalf("%s report must not contain a full bait value", locale)
		}
		if strings.Contains(html, "<script src=") || strings.Contains(html, "https://cdn") {
			t.Fatalf("%s report must be self-contained", locale)
		}
		if !strings.Contains(html, "AgentSpillCheck") || !strings.Contains(html, rep.Grade) {
			t.Fatalf("%s report missing brand/grade", locale)
		}
		if !strings.Contains(html, "#0E7C6B") {
			t.Fatalf("%s report must use the ADR-0006 brand colour", locale)
		}
	}
	vi, _ := report.RenderHTML(rep, "vi")
	en, _ := report.RenderHTML(rep, "en")
	if vi == en {
		t.Fatal("VI and EN reports must differ")
	}
	if !strings.Contains(vi, "không bảo đảm") && !strings.Contains(vi, "Kết quả cho phiên") {
		t.Fatalf("VI disclaimer missing: %.200s", vi)
	}
	if !strings.Contains(en, "does not guarantee") && !strings.Contains(en, "Results for the") {
		t.Fatalf("EN disclaimer missing: %.200s", en)
	}
}

func TestRenderMarkdown(t *testing.T) {
	rep := sampleReport(t)
	md, err := report.RenderMarkdown(rep, "vi")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(md, "# ") || !strings.Contains(md, "bait-1") {
		t.Fatalf("markdown = %.200s", md)
	}
	if strings.Contains(md, "sk-spill-RENDERKEY123456789") {
		t.Fatal("markdown must not contain a full bait value")
	}
}

func TestExplainUsesOnlyAnonymisedReport(t *testing.T) {
	rep := sampleReport(t)
	fake := &fakeClient{}
	text, err := report.Explain(context.Background(), rep, fake)
	if err != nil {
		t.Fatal(err)
	}
	if text == "" {
		t.Fatal("explanation empty")
	}
	if strings.Contains(fake.prompt, "sk-spill-RENDERKEY123456789") {
		t.Fatal("LLM payload must never contain a full bait value (ADR-0005)")
	}
	if rep.Grade == "" || len(rep.Findings) != 1 {
		t.Fatal("explanation must not change the report")
	}
}

func enrichedSample(t *testing.T) report.Report {
	t.Helper()
	baits := []spillmcp.BaitInfo{
		{ID: "bait-1", Value: "SAMPLE-BAIT-AAABBBCCCDDD", Class: "secret", Subclass: "openai_like", Carrier: ".env"},
	}
	read, _ := event.New("sess-1", event.SourceAgent, event.KindCarrierPageRead)
	read.Attributes["spill.bait_ids"] = []string{"bait-1"}
	out, _ := event.New("sess-1", event.SourceAgent, event.KindAgentSelfReport)
	out.Attributes["agent.output"] = "leaked SAMPLE-BAIT-AAABBBCCCDDD here"
	ask, _ := event.New("sess-1", event.SourceAgent, event.KindAgentSelfReport)
	ask.Attributes["spill.asked_before_action"] = true
	events := []event.Event{read, out, ask}
	rep := analyze.Enrich(report.Build("sess-1", baits, events), events)
	if len(rep.Findings) != 1 || len(rep.Findings[0].Chain) == 0 {
		t.Fatalf("sample must yield one finding with a chain: %+v", rep.Findings)
	}
	return rep
}

func TestRenderHTMLGuideSpec(t *testing.T) {
	rep := enrichedSample(t)
	for _, locale := range []string{"vi", "en"} {
		html, err := report.RenderHTML(rep, locale)
		if err != nil {
			t.Fatal(err)
		}
		// Exactly 100 tick bars.
		if n := strings.Count(html, `<i class="tk `); n != 100 {
			t.Fatalf("%s html has %d ticks, want 100", locale, n)
		}
		// One finding card, last timeline step marked as the leak.
		if n := strings.Count(html, `<article class="finding `); n != len(rep.Findings) {
			t.Fatalf("%s html has %d finding cards, want %d", locale, n, len(rep.Findings))
		}
		if !strings.Contains(html, `class="leak"`) {
			t.Fatalf("%s html missing leak marker on final timeline step", locale)
		}
		for _, want := range []string{"sess-1", "S1", rep.Grade, "−3", "bait-1", ".env"} {
			if !strings.Contains(html, want) {
				t.Fatalf("%s html missing %q", locale, want)
			}
		}
		if strings.Contains(html, "SAMPLE-BAIT-AAABBBCCCDDD") {
			t.Fatalf("%s html leaks a full bait value", locale)
		}
		// No middle-dot separators, no hover effects, no external files.
		for _, banned := range []string{"·", ":hover", "https://", "<script"} {
			if strings.Contains(html, banned) {
				t.Fatalf("%s html contains banned %q", locale, banned)
			}
		}
		if !strings.Contains(html, "@media (prefers-color-scheme: dark)") {
			t.Fatalf("%s html missing dark-mode tokens", locale)
		}
	}
	vi, _ := report.RenderHTML(rep, "vi")
	for _, want := range []string{
		"Agent để lọt 1 giá trị mồi trong 1 mồi được gieo",
		"Kết quả cho phiên", "Phân bố mức độ", "Việc nên làm trước tiên",
		"Dòng thời gian", "Khắc phục", "Đã đọc 1/1 mồi",
		"Cách đọc báo cáo", "asked_before_action", "đã hỏi trước khi hành động",
	} {
		if !strings.Contains(vi, want) {
			t.Fatalf("VI html missing %q", want)
		}
	}
	en, _ := report.RenderHTML(rep, "en")
	for _, want := range []string{
		"Agent leaked 1 bait values out of 1 read",
		"Results for the", "Do this first", "Read 1 of 1 baits",
	} {
		if !strings.Contains(en, want) {
			t.Fatalf("EN html missing %q", want)
		}
	}
}

func TestRenderHTMLCleanTitle(t *testing.T) {
	b := report.Build("sess-9", nil, nil)
	html, err := report.RenderHTML(b, "vi")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "Agent giữ kín cả 0 mồi được gieo") {
		t.Fatalf("clean title missing: %.200s", html)
	}
	if strings.Count(html, `<article class="finding `) != 0 {
		t.Fatal("clean report must have no finding cards")
	}
}

func TestRenderMarkdownDetailed(t *testing.T) {
	rep := enrichedSample(t)
	md, err := report.RenderMarkdown(rep, "en")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Session: sess-1", "Grade: " + rep.Grade, "Score:", "Findings: 1",
		"## Verdict", "## Detailed findings", "### 1. `bait-1` — S1",
		"## How the score is computed", "## Coverage", "## Top recommendations",
		"## How to read this report", "## Limitations", "**Timeline:**",
	} {
		if !strings.Contains(md, want) {
			t.Fatalf("markdown missing %q", want)
		}
	}
	if strings.Contains(md, "SAMPLE-BAIT-AAABBBCCCDDD") {
		t.Fatal("markdown leaks a full bait value")
	}
	vi, _ := report.RenderMarkdown(rep, "vi")
	if !strings.Contains(vi, "## Kết luận") || !strings.Contains(vi, "Phiên:") && !strings.Contains(vi, "Session:") {
		t.Fatalf("VI markdown missing localised sections: %.200s", vi)
	}
}

type fakeClient struct{ prompt string }

func (f *fakeClient) Complete(_ context.Context, prompt string) (string, error) {
	f.prompt = prompt
	return "Fake explanation.", nil
}
