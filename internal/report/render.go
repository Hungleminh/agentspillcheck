// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	"github.com/Hungleminh/agentspillcheck/internal/core/i18n"
)

// BrandColour is the ADR-0006 primary colour.
const BrandColour = "#0E7C6B"

// RenderHTML renders a self-contained localised HTML report. All
// agent-controlled values pass through html/template escaping.
func RenderHTML(rep Report, locale string) (string, error) {
	tr, err := i18n.New(locale)
	if err != nil {
		tr, _ = i18n.New("en")
	}
	loc := tr.Locale()
	data := map[string]any{
		"Lang": loc, "Brand": BrandColour, "Title": tr.T("report.title"),
		"Disclaimer": tr.T("report.disclaimer"), "GradeLabel": tr.T("report.grade"),
		"ScoreLabel": tr.T("report.score"), "Grade": rep.Grade, "Score": rep.Score,
		"Summary": rep.SummaryI18n[loc], "Bait": tr.T("report.bait"),
		"Exposure": tr.T("report.exposure"), "Destination": tr.T("report.destination"),
		"Severity": tr.T("report.severity"), "RootCause": tr.T("report.root_cause"),
		"ChainLabel": tr.T("report.chain"), "CoverageLabel": tr.T("report.coverage"),
		"Findings": rep.Findings, "Coverage": rep.Coverage, "Explanation": rep.Explanation,
	}
	var buf bytes.Buffer
	t2, err := template.New("r").Funcs(template.FuncMap{
		"join": func(parts []string, sep string) string { return strings.Join(parts, sep) },
	}).Parse(reportHTMLSource)
	if err != nil {
		return "", err
	}
	if err := t2.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

const reportHTMLSource = `<!DOCTYPE html>
<html lang="{{.Lang}}"><head><meta charset="utf-8">
<meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src 'unsafe-inline'; img-src data:">
<title>{{.Title}}</title>
<style>body{font-family:system-ui,sans-serif;margin:2rem;max-width:70rem}h1{color:{{.Brand}}}table{border-collapse:collapse}td,th{border:1px solid #ccc;padding:.35rem .6rem}.sev{font-weight:700}</style>
</head><body>
<h1>AgentSpillCheck</h1><h2>{{.Title}}</h2>
<p><em>{{.Disclaimer}}</em></p>
<p>{{.GradeLabel}}: <strong>{{.Grade}}</strong> · {{.ScoreLabel}}: <strong>{{.Score}}</strong></p>
<p>{{.Summary}}</p>
<table><tr><th>{{.Bait}}</th><th>{{.Exposure}}</th><th>{{.Destination}}</th><th>{{.Severity}}</th><th>{{.RootCause}}</th></tr>
{{range .Findings}}<tr><td>{{.BaitID}} ({{.MaskedValue}})</td><td>{{.ExposureType}}</td><td>{{.Destination}}</td><td class="sev">{{.Severity}}</td><td>{{join .RootCause ", "}}</td></tr>
{{end}}</table>
<h3>{{.ChainLabel}}</h3>
{{range .Findings}}<p>{{.BaitID}}: {{range .Chain}}{{.Kind}}@{{.Source}} → {{end}}</p>{{end}}
<h3>{{.CoverageLabel}}</h3><p>read={{len .Coverage.Read}} unread={{len .Coverage.Unread}}</p>
{{if .Explanation}}<h3>Explanation</h3><p>{{.Explanation}}</p>{{end}}
</body></html>`

// RenderMarkdown renders the Markdown report form.
func RenderMarkdown(rep Report, locale string) (string, error) {
	tr, err := i18n.New(locale)
	if err != nil {
		tr, _ = i18n.New("en")
	}
	loc := tr.Locale()
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n_%s_\n\n%s: **%s** · %s: **%d**\n\n%s\n\n",
		tr.T("report.title"), tr.T("report.disclaimer"),
		tr.T("report.grade"), rep.Grade, tr.T("report.score"), rep.Score,
		rep.SummaryI18n[loc])
	fmt.Fprintf(&b, "| %s | %s | %s | %s |\n|---|---|---|---|\n",
		tr.T("report.bait"), tr.T("report.exposure"), tr.T("report.destination"), tr.T("report.severity"))
	for _, f := range rep.Findings {
		fmt.Fprintf(&b, "| %s (%s) | %s | %s | %s |\n",
			mdEscape(f.BaitID), mdEscape(f.MaskedValue), mdEscape(f.ExposureType),
			mdEscape(f.Destination), mdEscape(f.Severity))
	}
	if rep.Explanation != "" {
		fmt.Fprintf(&b, "\n## Explanation\n\n%s\n", rep.Explanation)
	}
	return b.String(), nil
}

func mdEscape(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}

// Client is an LLM provider used only for explanations (ADR-0005).
type Client interface {
	Complete(ctx context.Context, prompt string) (string, error)
}

// Explain asks client to explain the anonymised report. The report
// itself is never modified; the payload contains masked values only.
func Explain(ctx context.Context, rep Report, client Client) (string, error) {
	anon := rep
	anon.Explanation = ""
	raw, err := json.Marshal(anon)
	if err != nil {
		return "", err
	}
	prompt := "Explain this anonymised agent data-leak report in plain language. " +
		"Do not invent findings. The score and grade are final.\n\n" + string(raw)
	return client.Complete(ctx, prompt)
}
