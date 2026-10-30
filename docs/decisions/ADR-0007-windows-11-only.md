# ADR-0007: Windows 11 only for v0.1 installers

- Status: Accepted
- Date: 2026-10-30
- Question: Q6 (docs/PLAN.md section 3)

## Context

PLAN 16.2 requires testing `install.bat` with and without Docker
Desktop; the supported Windows range had to be fixed first.

## Decision

v0.1 officially supports Windows 11 only. Windows 10 may run the
binary but is not tested or supported.

Confirmed by the project owner on 2026-10-30.

## Consequences

Installer tests and docs target Windows 11 (GitHub `windows-latest`).
Do not add Windows 10 workarounds without a new ADR.
