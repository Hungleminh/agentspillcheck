# ADR-0004: Persona names are random per session with an exclusion list

- Status: Accepted
- Date: 2026-10-30
- Question: Q8 (docs/PLAN.md section 3)

## Context

Personas need Vietnamese names that do not identify a real person
(PLAN 7.4). The owner named no specific person to avoid.

## Decision

Generate the full name randomly per session from Vietnamese family,
middle and given name dictionaries, skipping combinations on a small
exclusion list of well-known public figures. Demo mode uses one fixed
persona that is clearly labelled as fake.

Confirmed by the project owner on 2026-10-30.

## Consequences

Tests must assert generated personas are not on the exclusion list and
that two sessions with different seeds normally get different names.
No real person's name is committed as a persona.
