# ADR-0001: GitHub owner for module path

- Status: Accepted
- Date: 2026-10-30
- Question: Q1 (docs/PLAN.md section 3)

## Context

Go module path, README links, goreleaser and GHCR image name all embed the
GitHub owner: `github.com/<OWNER>/agentspillcheck`.

## Decision

Owner is `Hungleminh`.

- Go module: `github.com/Hungleminh/agentspillcheck`
- Repository: `github.com/Hungleminh/agentspillcheck`
- Docker image: `ghcr.io/hungleminh/agentspillcheck`

Confirmed by the project owner on 2026-10-30. The existing git remote
`https://github.com/Hungleminh/agentspillcheck` matches this decision.

## Consequences

Changing the owner later is a breaking change for the module path and
requires a new ADR.
