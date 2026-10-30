---
name: generate-report
description: Use when turning an AgentSpillCheck report.md into a standalone single-file HTML report, or when restyling static severity ladders, score bars, finding timelines, or coverage chips with CSS only and no JavaScript.
---

# Generate Report

## Overview

This skill converts a `report.md` produced by AgentSpillCheck into one
self-contained HTML file that keeps every number and stays readable
offline. Core principle: never invent numbers, never drop a section,
never print a real bait value (reports carry masked samples only).

## When to Use

- A `report.md` (or `report.json`) exists and a human needs a visual HTML version.
- Restyling severity, score, timeline, or coverage blocks with CSS only.
- Symptoms: hand-built report HTML with hardcoded colors, external
  fonts that break offline, JavaScript charts, missing findings.

When NOT to use: JSON/Markdown output is enough; the consumer is
another program; interactivity beyond anchor links or
`<details>` is required (that needs an app, not this static report).

## Input: report.md map

Read sections in order; every one below is always present.

| Section | Extract | Notes |
| --- | --- | --- |
| Italic line at top | Disclaimer sentence | Keep verbatim, right under the title |
| `Session`, `Grade`, `Score`, `Findings` | IDs, A–F, 0–100, count | Machine header, one field per line |
| Verdict | Grade, score, finding count | Cross-check against the header |
| Severity breakdown | S1–S4 counts with names | Always 4 levels, even when 0 |
| Detailed findings (N) | Per finding: bait id, severity, masked sample, bait type, carrier, exposure, destination, transport, encoding, cause, fix, timeline | Timeline items are `time action@place (evt-id)` |
| Scoring | Start 100, each minus/plus line (points, severity, bait, channel), final | Must add up to Score |
| Coverage | Read and unread bait lists | Read count may exceed findings |
| Top recommendation | One action sentence | Matches the heaviest finding's fix |
| How to read, Limits | Fixed paragraphs | Keep verbatim |

A timeline line `10:31:46 carrier.page_read@stage.drive
(evt-9d9b1341f686293f)` is time, event, place, event id. The last
event of a finding is the leak moment. A `—` destination means no
destination (leaked in place). If the card count differs from
`Findings`, stop and report the mismatch instead of fixing it silently.

## Non-negotiables

- One HTML file, inline `<style>`, no JavaScript, no external assets
  (no fonts, images, icon fonts). System font stack only so the file
  opens offline.
- All colors as CSS variables on `:root`, redefined under
  `@media (prefers-color-scheme: dark)` guarded by
  `:root:not([data-theme="light"])` and repeated under
  `:root[data-theme="dark"]`. No hardcoded colors in rules below.
- Sentence case everywhere. No all-caps labels, no `·` separators,
  no arrows on buttons, no hover effects on cards.
- One animation only: score ticks rising on load, lost ticks delayed
  0.35s, everything inside
  `@media (prefers-reduced-motion: no-preference)`.
- Zero-count severities still render, dimmed, so readers see S3/S4
  were tested.
- Small colored text must stay readable: mix severity colors toward
  the ink color (for example `color-mix(in srgb, var(--s1) 62%,
  var(--ink))`) instead of raw swatches on tint backgrounds.
- Keyboard: visible `:focus-visible` rings, semantic sections with
  `aria-labelledby`, decorative bars `aria-hidden="true"`.
- Print stylesheet: white background, no width cap,
  `break-inside: avoid` on cards, no tick animation.

## Quick Reference: tokens

| Token | Light | Dark | Use |
| --- | --- | --- | --- |
| `--paper` | `#F6F7F5` | `#141821` | Page background |
| `--panel` | `#FFFFFF` | `#1C2230` | Cards, boxes |
| `--ink` / `--ink-2` / `--ink-3` | `#1B222C` / `#4E5967` / `#8891A0` | `#EEF1F5` / `#B7C0CD` / `#7C8796` | Text tiers |
| `--line` / `--line-strong` | `#DDE1E6` / `#B9C0CA` | `#2B3342` / `#414B5E` | Borders, timeline axis |
| `--keep` | `#1B222C` | `#D9DEE6` | Kept score ticks |
| `--s1` / `--s2` / `--s3` / `--s4` | `#D4A017` / `#D8582C` / `#A2283A` / `#5A3C9A` | `#E4B53A` / `#E8714A` / `#D2546A` / `#9B82D6` | Severity (amber, red-orange, dark red, purple) |
| `--tint-s1` … `--tint-s4` | `#FBF3DC` `#FBE6DD` `#F6DFE2` `#E9E2F4` | `#35301D` `#3A261E` `#3A222A` `#2B2640` | Severity label backgrounds |
| `--grade` / `--grade-soft` | `#1F7A68` / `#E3F2EE` | `#3FB39A` / `#1D3A35` | Grade A/B accents |
| `--focus` | `#2A6AD6` | `#2A6AD6` | Keyboard focus ring |

Grade C redefines `--grade: #B07A12` / soft `#F7EDD3`; grade D
redefines `--grade: #A2283A` / soft `#F6DFE2` (dark variants follow
the same hues).

Type scale: page title `clamp(28px, 4.2vw, 40px)` weight 700;
h2 22px weight 600; body 16px/1.55; grade letter 84px; big numbers
40–44px. Content `max-width: 900px`, sections 56px apart,
breakpoints 760px and 640px, no horizontal scroll, long values get
`overflow-wrap: anywhere`.

## Layout order

Header (tool mark + session/date), one-sentence title from the data
(`Agent leaked <n> bait values out of <m> read`, or the kept-secret
variant when zero), disclaimer, hero (grade box 148px + score +
100-tick bar with taller every-10th ticks + kept/lost legend),
severity ladder (4 cells, 5px color edge, 44px counts, danger pips:
S1 lights one dot through S4 lighting four), one card per finding
(head: severity label + bait id + one-line description + penalty;
two-column facts and timeline; fix strip), scoring plus coverage
grid, top-recommendation box, how-to-read plus limits boxes, footer.

## Data mapping essentials

- Ticks: first `Score` bars keep-colored; the rest split per
  deduction line in order (for example 3 S1 + 3 S1 + 8 S2 fills bars
  87–92, 93–100). Bonus points never draw past 100; list them as
  text lines instead.
- Finding description from type, carrier, exposure: API-like key in
  its file printed verbatim into the answer; AWS-like key sent
  outward; fictional content sent outward. Unknown types fall back to
  the raw type name.
- Timeline verbs: `carrier.page_read` reads a page,
  `carrier.downloaded` downloads a file, `agent.self_report` prints
  out, `mail.sent` sends email, `http.request` sends HTTP,
  `form.submitted` submits a form; unknown codes stay verbatim.
  Places read `at / from / via / in <source>` by context. The final
  step carries the leak highlight.
- Coverage chips: leaked baits get a severity dot, safe ones a grade
  dot, unread ones a dashed border; the bar width is read/total.

## react-bits patterns, CSS-only

The react-bits library (animated React components) cannot run here:
its components need a JavaScript runtime, which this static report
forbids. Port only its visual language:

- Staggered entrance becomes the single tick-rise keyframe with the
  lost-tick delay; never animate every block.
- Spotlight/mouse-follow effects are dropped (they need cursor
  tracking); keep flat cards with tinted severity surfaces instead.
- Counters stay static numbers; progress becomes the tick bar and
  the coverage bar.
- Light/dark token pairs replace per-component themes.

## Common Mistakes

| Mistake | Fix |
| --- | --- |
| Web font link for offline file | System stack only; the file must open with no network |
| Hardcoded hex in a rule | Add a `--token`, define it in all three theme blocks |
| Raw severity color as small text | Mix toward ink (`color-mix`) to keep 4.5:1 contrast |
| Dropping zero-count severities | Render them dimmed; absence of evidence is information |
| Rewriting disclaimer/limits/method | Copy verbatim; only fix/remediation lines may be rephrased |
| Counts that disagree | Stop and report; never adjust numbers to fit |

## Acceptance checklist

Ticks lost equal 100 minus Score and split per deduction line;
deductions add up to Score; card count equals Findings and the S1–S4
sum; every card has its facts, fix strip, and full timeline with the
leak marked; masked samples only, nothing outside the md; title is a
conclusion sentence with numbers; no banned separators or hover;
correct at 360/768/1280px with no sideways scroll; dark mode swaps
every surface; keyboard focus visible; reduced-motion disables motion;
prints to PDF without split cards; opens with no network.
