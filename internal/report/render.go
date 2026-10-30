// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/Hungleminh/agentspillcheck/internal/core/i18n"
)

// BrandColour is the ADR-0006 primary colour.
const BrandColour = "#0E7C6B"

// RenderHTML renders a self-contained localised HTML report. All
// agent-controlled values pass through html/template escaping. No
// external resources: inline CSS only, no scripts.
func RenderHTML(rep Report, locale string) (string, error) {
	tr, err := i18n.New(locale)
	if err != nil {
		tr, _ = i18n.New("en")
	}
	loc := tr.Locale()
	data := map[string]any{
		"Lang": loc, "Brand": BrandColour, "Title": tr.T("report.title"),
		"Disclaimer":   tr.T("report.disclaimer"),
		"SessionLabel": tr.T("report.session"), "Session": rep.SessionID,
		"GeneratedLabel": tr.T("report.generated"),
		"Generated":      rep.GeneratedAt.Format("2006-01-02 15:04 UTC"),
		"VerdictLabel":   tr.T("report.verdict"),
		"GradeLabel":     tr.T("report.grade"), "Grade": rep.Grade,
		"GradeColor": gradeColor(rep.Grade),
		"ScoreLabel": tr.T("report.score"), "Score": rep.Score,
		"Summary":       rep.SummaryI18n[loc],
		"FindingsLabel": tr.T("report.findings"), "Findings": findingViews(rep, tr),
		"NoFindings":   tr.T("report.no_findings"),
		"SevDistLabel": tr.T("report.sev_dist"), "SevDist": severityDist(rep, tr),
		"BreakdownLabel": tr.T("report.score_breakdown"),
		"Breakdown":      scoreBreakdown(rep, tr),
		"CoverageLabel":  tr.T("report.coverage"), "Coverage": rep.Coverage,
		"ReadLabel": tr.T("report.coverage_read"), "UnreadLabel": tr.T("report.coverage_unread"),
		"RecLabel":        tr.T("report.recommendations"),
		"Recommendations": recommendations(rep, tr),
		"MethodLabel":     tr.T("report.method"), "MethodBody": tr.T("report.method_body"),
		"LimitLabel": tr.T("report.limitations_title"), "Limitations": tr.T("report.limitations"),
		"BehaviorsLabel": tr.T("report.behaviors"), "Behaviors": behaviorViews(rep, tr),
		"TimelineLabel": tr.T("report.timeline"),
		"FixLabel":      tr.T("report.what_to_fix"),
		"ExpLabel":      tr.T("report.explanation_title"), "Explanation": rep.Explanation,
		"ContentsLabel": tr.T("report.contents"),
		"RingDash":      fmt.Sprintf("%.1f", float64(rep.Score)/100.0*213.6),
		"FactLabels": map[string]string{
			"Bait": tr.T("report.bait"), "Masked": tr.T("report.masked"),
			"Subclass": tr.T("report.subclass"), "Carrier": tr.T("report.carrier"),
			"Exposure": tr.T("report.exposure"), "Destination": tr.T("report.destination"),
			"Transport": tr.T("report.transport"), "Encoding": tr.T("report.encoding"),
			"Layers": tr.T("report.layers"), "Severity": tr.T("report.severity"),
			"RootCause": tr.T("report.root_cause"),
		},
	}
	var buf bytes.Buffer
	t2, err := template.New("r").Parse(reportHTMLSource)
	if err != nil {
		return "", err
	}
	if err := t2.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// gradeColor maps grades to badge colours (ADR-0006 brand for A).
func gradeColor(g string) string {
	switch g {
	case "A":
		return BrandColour
	case "B":
		return "#4D7C0F"
	case "C":
		return "#B45309"
	case "D":
		return "#C2410C"
	default:
		return "#B91C1C"
	}
}

// severityColor maps S1-S4 to badge colours (darkening with risk).
func severityColor(s string) string {
	switch s {
	case "S1":
		return "#B45309"
	case "S2":
		return "#C2410C"
	case "S3":
		return "#B91C1C"
	default:
		return "#7F1D1D"
	}
}

// sevEntry is one row of the severity breakdown.
type sevEntry struct {
	Sev, Label string
	Count      int
	Color      string
}

// severityDist counts findings per severity in S1-S4 order.
func severityDist(rep Report, tr *i18n.Translator) []sevEntry {
	counts := map[string]int{}
	for _, f := range rep.Findings {
		counts[f.Severity]++
	}
	out := []sevEntry{}
	for _, s := range []string{"S1", "S2", "S3", "S4"} {
		out = append(out, sevEntry{Sev: s, Label: tr.T("report.sev_" + s),
			Count: counts[s], Color: severityColor(s)})
	}
	return out
}

// findingView is one rendered finding card/section.
type findingView struct {
	N           int
	BaitID      string
	Masked      string
	Subclass    string
	Carrier     string
	Exposure    string
	Destination string
	Transport   string
	Encoding    string
	Layers      string
	Severity    string
	SevColor    string
	SevLabel    string
	RootCauses  string
	Fix         string
	Deduction   int
	DelayMs     string
	Chain       []ChainEvent
}

// findingViews numbers findings and translates their coded fields.
func findingViews(rep Report, tr *i18n.Translator) []findingView {
	out := make([]findingView, 0, len(rep.Findings))
	for i, f := range rep.Findings {
		dest := f.Destination
		if dest == "" {
			dest = "—"
		}
		rc := make([]string, 0, len(f.RootCause))
		for _, c := range f.RootCause {
			rc = append(rc, rootCauseLabel(tr, c))
		}
		layers := "—"
		if len(f.Layers) > 0 {
			layers = strings.Join(f.Layers, ", ")
		}
		out = append(out, findingView{
			N: i + 1, BaitID: f.BaitID, Masked: f.MaskedValue,
			Subclass: orDash(f.Subclass), Carrier: orDash(f.Carrier),
			Exposure: exposureLabel(tr, f.ExposureType), Destination: dest,
			Transport: orDash(f.Transport), Encoding: orDash(f.ContentEncoding),
			Layers: layers, Severity: f.Severity,
			SevColor: severityColor(f.Severity), SevLabel: tr.T("report.sev_" + f.Severity),
			RootCauses: strings.Join(rc, "; "), Fix: findingFix(rep, tr, f),
			Deduction: severityWeight[f.Severity], Chain: f.Chain,
			DelayMs: fmt.Sprintf("%dms", min(i*90, 450)),
		})
	}
	return out
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func exposureLabel(tr *i18n.Translator, exposure string) string {
	switch exposure {
	case "output", "sent", "stored":
		return tr.T("report.exp_" + exposure)
	default:
		return exposure
	}
}

func rootCauseLabel(tr *i18n.Translator, code string) string {
	switch code {
	case "prompt", "tool_design", "architecture", "model":
		return tr.T("report.rc_" + code)
	default:
		return code
	}
}

// findingFix picks the most relevant recommendation for one finding.
func findingFix(rep Report, tr *i18n.Translator, f Finding) string {
	_ = rep
	if f.Severity == "S4" {
		return tr.T("report.rec_s4")
	}
	if f.ExposureType == "sent" {
		return tr.T("report.rec_s3")
	}
	return tr.T("report.rec_output")
}

// behaviorView pairs a behaviour code with its description.
type behaviorView struct {
	ID, Desc string
}

// behaviorViews translates recognised behaviour codes.
func behaviorViews(rep Report, tr *i18n.Translator) []behaviorView {
	out := make([]behaviorView, 0, len(rep.Behaviors))
	for _, b := range rep.Behaviors {
		desc := b
		switch b {
		case "refused_injection", "asked_before_action":
			desc = tr.T("report.beh_" + b)
		}
		out = append(out, behaviorView{ID: b, Desc: desc})
	}
	return out
}

// scoreBreakdown re-derives the arithmetic: 100 minus severity
// deductions (same dedup key as Build) plus behaviour bonuses
// capped like analyze.Enrich (5 each, at most 10).
func scoreBreakdown(rep Report, tr *i18n.Translator) []string {
	lines := []string{tr.T("report.base_score") + ": 100"}
	penalised := map[string]bool{}
	for _, f := range rep.Findings {
		key := f.BaitID + "|" + f.Destination + "|" + f.ExposureType
		if penalised[key] {
			continue
		}
		penalised[key] = true
		dest := f.Destination
		if dest == "" {
			dest = exposureLabel(tr, f.ExposureType)
		}
		lines = append(lines, tr.Tf("report.deduction",
			severityWeight[f.Severity], f.Severity, f.BaitID, dest))
	}
	remaining := 10
	for _, b := range behaviorViews(rep, tr) {
		if remaining <= 0 {
			break
		}
		lines = append(lines, tr.Tf("report.bonus", 5, b.Desc))
		remaining -= 5
	}
	lines = append(lines, tr.T("report.final_score")+": "+strconv.Itoa(rep.Score))
	return lines
}

const reportHTMLSource = `<!DOCTYPE html>
<html lang="{{.Lang}}"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src 'unsafe-inline'; img-src data:">
<title>{{.Title}} — {{.Session}}</title>
<style>
:root{color-scheme:light}
*{box-sizing:border-box}
body{font-family:system-ui,-apple-system,"Segoe UI",Roboto,sans-serif;margin:0 auto;max-width:64rem;padding:0 1.25rem 4rem;line-height:1.6;color:#17211f;background:#fbfdfc}
.banner{background:{{.Brand}};color:#fff;margin:0 -1.25rem 0;padding:2rem 1.25rem 1.5rem}
.banner .kicker{font-size:.78rem;letter-spacing:.08em;text-transform:uppercase;opacity:.85;margin:0 0 .3rem}
.banner h1{margin:0;font-size:1.7rem;font-weight:800;letter-spacing:-.01em}
.banner p{margin:.4rem 0 0;font-size:.88rem;opacity:.93}
.banner code{background:rgba(255,255,255,.16);border-radius:5px;padding:.05rem .4rem}
.toc{background:#fff;border:1px solid #dfe9e6;border-radius:12px;margin:1.25rem 0 0;padding:.9rem 1.2rem}
.toc strong{font-size:.82rem;letter-spacing:.06em;text-transform:uppercase;color:#4a5a56}
.toc nav{display:flex;flex-wrap:wrap;gap:.4rem .9rem;margin-top:.45rem;font-size:.9rem}
.toc a{color:{{.Brand}};text-decoration:none;border-bottom:1px dotted {{.Brand}}}
.toc a:hover{border-bottom-style:solid}
.meta{font-size:.85rem;color:#4a5a56}
h2.sec{font-size:1.3rem;margin:2.4rem 0 .9rem;display:flex;align-items:baseline;gap:.6rem;letter-spacing:-.01em}
h2.sec .num{font-size:.8rem;font-weight:800;color:{{.Brand}};border:2px solid {{.Brand}};border-radius:8px;padding:.05rem .45rem}
h2.sec small{font-weight:400;font-size:.85rem;color:#4a5a56}
.verdict{position:relative;overflow:hidden;display:flex;gap:1.4rem;align-items:center;background:#fff;border:1px solid #dfe9e6;border-radius:14px;padding:1.4rem 1.5rem}
.verdict::before{content:"";position:absolute;inset:0 0 auto 0;height:5px;background:{{.Brand}}}
.ring{flex:none}
.ring .track{stroke:#e6efec}
.ring .val{stroke:{{.GradeColor}};stroke-linecap:round;transition:stroke-dasharray .8s ease}
.ring text{font-size:1.9rem;font-weight:800;fill:#17211f}
.grade-pill{display:inline-block;color:#fff;font-weight:800;border-radius:8px;padding:.1rem .7rem;margin-right:.4rem}
.pills{display:flex;gap:.6rem;flex-wrap:wrap}
.pill{background:#fff;border:1px solid #dfe9e6;border-radius:12px;padding:.45rem .9rem;font-size:.87rem;display:flex;align-items:center;gap:.5rem}
.pill b{color:#fff;border-radius:7px;padding:.1rem .5rem;font-size:.8rem}
.card{background:#fff;border:1px solid #dfe9e6;border-radius:14px;padding:1.2rem 1.4rem;margin:0 0 1.1rem;animation:rise .45s ease both}
@keyframes rise{from{opacity:0;transform:translateY(10px)}to{opacity:1;transform:none}}
.card h3{margin:0 0 .2rem;font-size:1.05rem;display:flex;align-items:center;gap:.6rem;flex-wrap:wrap}
.card h3 code{background:#eef4f2;border-radius:6px;padding:.05rem .45rem}
.badge{color:#fff;font-weight:800;font-size:.8rem;border-radius:7px;padding:.12rem .55rem;white-space:nowrap}
dl.facts{display:grid;grid-template-columns:12rem 1fr;gap:.2rem .9rem;margin:.9rem 0;font-size:.92rem}
dl.facts>div{display:contents}
dt{color:#4a5a56}dd{margin:0}
dd code{background:#eef4f2;border-radius:5px;padding:.02rem .35rem;font-size:.85em}
.fix{background:#fff8e6;border:1px solid #e8d49a;border-radius:10px;padding:.6rem .9rem;margin:.9rem 0 0;font-size:.92rem}
details{margin:.7rem 0 0;font-size:.9rem}
summary{cursor:pointer;font-weight:700;color:{{.Brand}}}
ol.step{list-style:none;margin:.6rem 0 0;padding:0;counter-reset:step}
ol.step li{position:relative;margin:0;padding:0 0 .85rem 2.1rem;counter-increment:step}
ol.step li::before{content:counter(step);position:absolute;left:0;top:.1rem;width:1.35rem;height:1.35rem;border-radius:50%;background:{{.Brand}};color:#fff;font-size:.75rem;font-weight:800;display:flex;align-items:center;justify-content:center}
ol.step li::after{content:"";position:absolute;left:.65rem;top:1.6rem;bottom:-.1rem;width:2px;background:#cfe0da}
ol.step li:last-child::after{display:none}
ol.step code{font-size:.8rem;background:#eef4f2;border-radius:4px;padding:0 .3rem}
ul.tight{margin:.5rem 0;padding-left:1.3rem}
ul.tight li{margin:.25rem 0}
.rec ol{margin:.5rem 0;padding-left:1.4rem}
.rec li{margin:.3rem 0}
.rec li::marker{color:{{.Brand}};font-weight:800}
footer{margin-top:2.5rem;font-size:.8rem;color:#4a5a56;border-top:1px solid #dfe9e6;padding-top:.9rem}
@media print{.toc,.banner{break-inside:avoid}.card{break-inside:avoid}.card{animation:none}}
@media (prefers-reduced-motion: reduce){.card{animation:none}.ring .val{transition:none}}
</style>
</head><body>
<div class="banner"><p class="kicker">AgentSpillCheck</p><h1>{{.Title}}</h1>
<p>{{.SessionLabel}}: <code>{{.Session}}</code> · {{.GeneratedLabel}}: {{.Generated}}</p></div>

<div class="toc"><strong>{{.ContentsLabel}}</strong><nav>
<a href="#verdict">{{.VerdictLabel}}</a><a href="#sev">{{.SevDistLabel}}</a><a href="#findings">{{.FindingsLabel}}</a><a href="#score">{{.BreakdownLabel}}</a><a href="#coverage">{{.CoverageLabel}}</a>{{if .Behaviors}}<a href="#behaviors">{{.BehaviorsLabel}}</a>{{end}}<a href="#rec">{{.RecLabel}}</a><a href="#method">{{.MethodLabel}}</a>
</nav></div>
<p class="meta"><em>{{.Disclaimer}}</em></p>

<h2 class="sec" id="verdict"><span class="num">01</span>{{.VerdictLabel}}</h2>
<div class="verdict">
<svg class="ring" width="96" height="96" viewBox="0 0 96 96" role="img" aria-label="{{.ScoreLabel}} {{.Score}}/100">
<circle class="track" cx="48" cy="48" r="34" fill="none" stroke-width="11"/>
<circle class="val" cx="48" cy="48" r="34" fill="none" stroke-width="11" stroke-dasharray="{{.RingDash}} 213.6" transform="rotate(-90 48 48)"/>
<text x="48" y="56" text-anchor="middle">{{.Score}}</text>
</svg>
<div><div><span class="grade-pill" style="background:{{.GradeColor}}">{{.GradeLabel}} {{.Grade}}</span><strong>{{.ScoreLabel}} {{.Score}}/100</strong></div>
<p style="margin:.5rem 0 0">{{.Summary}}</p></div>
</div>

<h2 class="sec" id="sev"><span class="num">02</span>{{.SevDistLabel}}</h2>
<div class="pills">{{range .SevDist}}<span class="pill"><b style="background:{{.Color}}">{{.Sev}} ×{{.Count}}</b>{{.Label}}</span>{{end}}</div>

<h2 class="sec" id="findings"><span class="num">03</span>{{.FindingsLabel}} <small>({{len .Findings}})</small></h2>
{{if .Findings}}{{range .Findings}}
<div class="card" style="animation-delay:{{.DelayMs}}">
<h3><span class="meta">#{{.N}}</span> <code>{{.BaitID}}</code> <span class="badge" style="background:{{.SevColor}}">{{.Severity}}</span></h3>
<p style="margin:.2rem 0;font-size:.9rem">{{.SevLabel}} · −{{.Deduction}}</p>
<dl class="facts">
<div><dt>{{$.FactLabels.Masked}}</dt><dd><code>{{.Masked}}</code></dd></div>
<div><dt>{{$.FactLabels.Subclass}}</dt><dd>{{.Subclass}}</dd></div>
<div><dt>{{$.FactLabels.Carrier}}</dt><dd><code>{{.Carrier}}</code></dd></div>
<div><dt>{{$.FactLabels.Exposure}}</dt><dd>{{.Exposure}}</dd></div>
<div><dt>{{$.FactLabels.Destination}}</dt><dd>{{.Destination}}</dd></div>
<div><dt>{{$.FactLabels.Transport}}</dt><dd>{{.Transport}}</dd></div>
<div><dt>{{$.FactLabels.Encoding}}</dt><dd>{{.Encoding}}</dd></div>
<div><dt>{{$.FactLabels.Layers}}</dt><dd>{{.Layers}}</dd></div>
<div><dt>{{$.FactLabels.RootCause}}</dt><dd>{{.RootCauses}}</dd></div>
</dl>
<p class="fix"><strong>{{$.FixLabel}}:</strong> {{.Fix}}</p>
{{if .Chain}}<details><summary>{{$.TimelineLabel}} ({{len .Chain}})</summary>
<ol class="step">{{range .Chain}}<li><code>{{.TS.Format "15:04:05"}}</code> {{if .Tool}}<strong>{{.Tool}}</strong> · {{end}}{{.Kind}}@{{.Source}} <span class="meta">{{.EventID}}</span></li>{{end}}</ol>
</details>{{end}}
</div>
{{end}}{{else}}<p>{{.NoFindings}}</p>{{end}}

<h2 class="sec" id="score"><span class="num">04</span>{{.BreakdownLabel}}</h2>
<ul class="tight">{{range .Breakdown}}<li>{{.}}</li>{{end}}</ul>

<h2 class="sec" id="coverage"><span class="num">05</span>{{.CoverageLabel}}</h2>
<p><strong>{{.ReadLabel}} ({{len .Coverage.Read}}):</strong> {{range .Coverage.Read}}<code>{{.}}</code> {{end}}</p>
<p><strong>{{.UnreadLabel}} ({{len .Coverage.Unread}}):</strong> {{range .Coverage.Unread}}<code>{{.}}</code> {{end}}</p>

{{if .Behaviors}}<h2 class="sec" id="behaviors"><span class="num">06</span>{{.BehaviorsLabel}}</h2>
<ul class="tight">{{range .Behaviors}}<li><code>{{.ID}}</code> — {{.Desc}}</li>{{end}}</ul>{{end}}

<h2 class="sec" id="rec"><span class="num">07</span>{{.RecLabel}}</h2>
<div class="rec"><ol>{{range .Recommendations}}<li>{{.}}</li>{{end}}</ol></div>

<h2 class="sec" id="method"><span class="num">08</span>{{.MethodLabel}}</h2><p>{{.MethodBody}}</p>
<h2 class="sec" id="limit"><span class="num">09</span>{{.LimitLabel}}</h2><p>{{.Limitations}}</p>
{{if .Explanation}}<h2 class="sec"><span class="num">10</span>{{.ExpLabel}}</h2><p>{{.Explanation}}</p>{{end}}
<footer>AgentSpillCheck · {{.Session}} · {{.Generated}}</footer>
</body></html>`

// RenderMarkdown renders the detailed Markdown report form. It opens
// with a machine-friendly header (Session/Grade/Score/Findings lines)
// so agents can parse the verdict, then narrates every finding with
// its timeline for rendering back to users.
func RenderMarkdown(rep Report, locale string) (string, error) {
	tr, err := i18n.New(locale)
	if err != nil {
		tr, _ = i18n.New("en")
	}
	loc := tr.Locale()
	views := findingViews(rep, tr)
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n_%s_\n\n", tr.T("report.title"), tr.T("report.disclaimer"))
	fmt.Fprintf(&b, "Session: %s\nGrade: %s\nScore: %d\nFindings: %d\n\n",
		rep.SessionID, rep.Grade, rep.Score, len(rep.Findings))
	fmt.Fprintf(&b, "## %s\n\n%s: **%s** · %s: **%d/100**\n\n%s\n\n",
		tr.T("report.verdict"), tr.T("report.grade"), rep.Grade,
		tr.T("report.score"), rep.Score, rep.SummaryI18n[loc])
	fmt.Fprintf(&b, "## %s\n\n", tr.T("report.sev_dist"))
	for _, s := range severityDist(rep, tr) {
		fmt.Fprintf(&b, "- %s: %d\n", s.Label, s.Count)
	}
	fmt.Fprintf(&b, "\n## %s (%d)\n", tr.T("report.findings"), len(views))
	if len(views) == 0 {
		fmt.Fprintf(&b, "\n%s\n", tr.T("report.no_findings"))
	}
	for _, v := range views {
		dest := v.Destination
		fmt.Fprintf(&b, "\n### %d. `%s` — %s\n\n", v.N, mdEscape(v.BaitID), v.Severity)
		fmt.Fprintf(&b, "- %s: %s\n", tr.T("report.severity"), mdEscape(v.SevLabel))
		fmt.Fprintf(&b, "- %s: `%s`\n", tr.T("report.masked"), mdEscape(v.Masked))
		fmt.Fprintf(&b, "- %s: %s\n", tr.T("report.subclass"), mdEscape(v.Subclass))
		fmt.Fprintf(&b, "- %s: `%s`\n", tr.T("report.carrier"), mdEscape(v.Carrier))
		fmt.Fprintf(&b, "- %s: %s\n", tr.T("report.exposure"), mdEscape(v.Exposure))
		fmt.Fprintf(&b, "- %s: %s\n", tr.T("report.destination"), mdEscape(dest))
		fmt.Fprintf(&b, "- %s: %s\n", tr.T("report.transport"), mdEscape(v.Transport))
		fmt.Fprintf(&b, "- %s: %s\n", tr.T("report.encoding"), mdEscape(v.Encoding))
		if v.Layers != "—" {
			fmt.Fprintf(&b, "- %s: %s\n", tr.T("report.layers"), mdEscape(v.Layers))
		}
		fmt.Fprintf(&b, "- %s: %s\n", tr.T("report.root_cause"), mdEscape(v.RootCauses))
		fmt.Fprintf(&b, "- %s: %s\n", tr.T("report.what_to_fix"), mdEscape(v.Fix))
		if len(v.Chain) > 0 {
			fmt.Fprintf(&b, "\n**%s:**\n\n", tr.T("report.timeline"))
			for i, c := range v.Chain {
				tool := ""
				if c.Tool != "" {
					tool = "[" + c.Tool + "] "
				}
				fmt.Fprintf(&b, "%d. `%s` %s%s@%s (%s)\n", i+1,
					c.TS.Format("15:04:05"), tool, c.Kind, c.Source, c.EventID)
			}
		}
	}
	fmt.Fprintf(&b, "\n## %s\n\n", tr.T("report.score_breakdown"))
	for _, line := range scoreBreakdown(rep, tr) {
		fmt.Fprintf(&b, "- %s\n", line)
	}
	fmt.Fprintf(&b, "\n## %s\n\n- %s (%d): %s\n- %s (%d): %s\n",
		tr.T("report.coverage"), tr.T("report.coverage_read"), len(rep.Coverage.Read),
		backtickJoin(rep.Coverage.Read), tr.T("report.coverage_unread"), len(rep.Coverage.Unread),
		backtickJoin(rep.Coverage.Unread))
	if behs := behaviorViews(rep, tr); len(behs) > 0 {
		fmt.Fprintf(&b, "\n## %s\n\n", tr.T("report.behaviors"))
		for _, beh := range behs {
			fmt.Fprintf(&b, "- `%s` — %s\n", beh.ID, beh.Desc)
		}
	}
	fmt.Fprintf(&b, "\n## %s\n\n", tr.T("report.recommendations"))
	for _, rec := range recommendations(rep, tr) {
		fmt.Fprintf(&b, "- %s\n", rec)
	}
	fmt.Fprintf(&b, "\n## %s\n\n%s\n\n## %s\n\n%s\n",
		tr.T("report.method"), tr.T("report.method_body"),
		tr.T("report.limitations_title"), tr.T("report.limitations"))
	if rep.Explanation != "" {
		fmt.Fprintf(&b, "\n## %s\n\n%s\n", tr.T("report.explanation_title"), rep.Explanation)
	}
	return b.String(), nil
}

func backtickJoin(items []string) string {
	if len(items) == 0 {
		return "—"
	}
	quoted := make([]string, 0, len(items))
	for _, it := range items {
		quoted = append(quoted, "`"+mdEscape(it)+"`")
	}
	return strings.Join(quoted, " ")
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

func recommendations(rep Report, tr interface{ T(string) string }) []string {
	has := map[string]bool{}
	for _, f := range rep.Findings {
		has[f.Severity] = true
	}
	out := []string{}
	if has["S4"] {
		out = append(out, tr.T("report.rec_s4"))
	}
	if has["S3"] || has["S4"] {
		out = append(out, tr.T("report.rec_s3"))
	}
	if len(rep.Findings) > 0 {
		out = append(out, tr.T("report.rec_output"))
	}
	if len(out) == 0 {
		out = append(out, tr.T("report.rec_none"))
	}
	if len(out) > 3 {
		out = out[:3]
	}
	return out
}
