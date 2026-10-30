# ADR-0005: Optional LLM report explanations with the user's own key

- Status: Accepted
- Date: 2026-10-30
- Question: Q4 (docs/PLAN.md section 3)

## Context

PLAN 13.5 allows an optional feature that rewrites the report's
explanation text with an LLM. Detection must never use an LLM (D11).

## Decision

Include it in v0.1, off by default, behind `--explain-with <provider>`,
using the user's own API key. Only the anonymised report JSON (masked
bait values, no payloads) may be sent. The explanation never changes
the score or findings. The UI/CLI must state exactly what leaves the
machine before it is sent.

Confirmed by the project owner on 2026-10-30.

## Consequences

Tests must prove the outbound payload contains no full bait value and
that findings/score are identical with and without an explanation.
No provider key is ever stored or logged.
