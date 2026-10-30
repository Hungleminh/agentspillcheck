# Progress

Updated at the end of every working session (PLAN 0.1.8).

## Done

- 2026-10-30: Read `docs/PLAN.md` in full. Owner answered Q1 and Q3.
- 2026-10-30: ADR-0001 (GitHub owner `Hungleminh`), ADR-0002 (ports 7700-7709).
- 2026-10-30: Step 1 of checklist 24.6 in progress on `feat/bootstrap-repo`:
  `docs/PLAN.md`, `NOTICE`, `SECURITY.md`, `AGENTS.md`, `.gitleaks.toml`,
  hardened `.gitignore`, Dependabot config.

## In progress

- Step 1 PR (open): https://github.com/Hungleminh/agentspillcheck/pull/1
  (`feat/bootstrap-repo` -> `main`).
- Step 4 PR (open, stacked): https://github.com/Hungleminh/agentspillcheck/pull/4
  (`feat/stage-drive-webmail` -> `feat/bait-generator`).
- Step 3 PR (open, stacked): https://github.com/Hungleminh/agentspillcheck/pull/3
  (`feat/bait-generator` -> `feat/cli-skeleton`).
- Step 2 PR (open, stacked): https://github.com/Hungleminh/agentspillcheck/pull/2
  (`feat/cli-skeleton` -> `feat/bootstrap-repo`; retarget to `main`
  after PR #1 merges).
- Step 2 implemented on `feat/cli-skeleton`: Go module
  `github.com/Hungleminh/agentspillcheck` (go 1.23), cobra CLI
  (`version`, `doctor`, `up`), i18n VI/EN, empty dashboard, `ci.yml`.
  Local evidence 2026-10-30: `gofmt -l` clean, `go vet ./...`,
  `go test ./... -count=1` (4 packages ok), `go build` ok.
  Sandbox blocks TCP `listen`, so a live `up` smoke test could not run
  here; the dashboard handler is covered by `httptest` tests instead.

## Next

- Step 5 of checklist 24.6: MCP server with spill_* and stage_* tools.

- 2026-10-30: PR #3 CI confirmed terminal: all 4 check-runs success.
- 2026-10-30: Step 4 on `feat/stage-drive-webmail`: `internal/event`
  (OTel-aligned schema, bus), `internal/store` (SQLite WAL via a
  pure-Go driver, append-only events), `internal/stage/drive` and
  `internal/stage/webmail` (JSON API + HTML, per-session namespaces,
  paginated reads, bait-tagged events, HTML escaping, body limits).
  Wiring bus events into the store happens when sessions are created
  (core session step); the stage emits to the bus today.
  Local evidence: `go vet ./...`, `go test ./... -count=1`
  (10 packages ok), `go build ./...`.

- 2026-10-30: Owner confirmed secret scanning + push protection and
  branch protection on `main` are enabled in GitHub Settings.
- 2026-10-30: Owner answered Q7/Q8; recorded in ADR-0003 (card bait:
  custom range failing Luhn) and ADR-0004 (persona names: random per
  session with an exclusion list).
- 2026-10-30: Step 3 on `feat/bait-generator`: `internal/bait`
  (4 bait groups, identifier embedding, fingerprints, JWT, card fails
  Luhn) and `internal/core/persona`. Local evidence: `go vet ./...`,
  `go test ./... -count=1` (6 packages ok), `go build ./...`.

## Waiting on the project owner

- GitHub repo settings that cannot be set from code: enable secret
  scanning + push protection, and branch protection on `main`
  (Settings -> Code security / Branches). PRs #1/#2 were opened via the
  GitHub API because no `gh` CLI is installed in this environment.
- Q2, Q4-Q10 remain open; each will be asked when its step is reached
  (Q7/Q8 before bait/persona, Q4/Q9 before HTML report, Q6 before
  installers, Q5 before video, Q2 before v0.2, Q10 at CI hardening).
