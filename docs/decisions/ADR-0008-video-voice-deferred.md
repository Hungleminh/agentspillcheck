# ADR-0008: Video voice (Q5) deferred by the owner

- Status: Deferred
- Date: 2026-10-30
- Question: Q5 (docs/PLAN.md section 3)

## Context

PLAN 21 needs a voice/subtitle choice for the 150-180s intro video.
The owner rejected both proposed options (Vietnamese voice with
English subtitles, and the reverse) and then chose to defer.

## Decision

Q5 is intentionally undecided. The video specification in
`docs/video/` stays in draft, no voice track is produced for v0.1,
and `spillcheck demo` (the recordable demo path) does not depend on
this decision.

Recorded on 2026-10-30.

## Consequences

Nothing in code or the release is blocked. The video agent must ask
again before producing any voice track.
