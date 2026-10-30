# ADR-0002: Default port range 7700-7709

- Status: Accepted
- Date: 2026-10-30
- Question: Q3 (docs/PLAN.md section 3)

## Context

PLAN section 8.1 assigns one port per stage service. The proposed default
range is 7700-7709, with the dashboard on 7700.

## Decision

Keep `port_base = 7700`, ports 7700-7709. The range does not conflict with
the project owner's infrastructure.

The base remains overridable via `spillcheck up --port-base`,
`SPILLCHECK_PORT_BASE`, and `port_base` in `~/.spillcheck/config.yaml`.

Confirmed by the project owner on 2026-10-30.

## Consequences

Defaults in config, CLI help, installers and docs use 7700-7709.
`spillcheck doctor` must report a clear error when a port is taken.
