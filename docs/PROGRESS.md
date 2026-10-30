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
- Step 9 PR (open, stacked): https://github.com/Hungleminh/agentspillcheck/pull/9
  (`feat/connectors-api` -> `feat/analyzer-reports`).
- Step 8 PR (open, stacked): https://github.com/Hungleminh/agentspillcheck/pull/8
  (`feat/analyzer-reports` -> `feat/stage-injection-sinks`).
- Step 7 PR (open, stacked): https://github.com/Hungleminh/agentspillcheck/pull/7
  (`feat/stage-injection-sinks` -> `feat/scenarios-detectors`).
- Step 6 PR (open, stacked): https://github.com/Hungleminh/agentspillcheck/pull/6
  (`feat/scenarios-detectors` -> `feat/mcp-server`).
- Step 5 PR (open, stacked): https://github.com/Hungleminh/agentspillcheck/pull/5
  (`feat/mcp-server` -> `feat/stage-drive-webmail`).
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

- Step 10 of checklist 24.6: manual tests (PLAN 18.4), release
  `v0.1.0`. Still open: Q2 (stage domain, before v0.2), Q10
  (Dependabot auto-merge, at CI hardening). Manual Claude Desktop /
  Windows 11 VM / macOS tests need real environments and remain
  unverified; tagging `v0.1.0-rc1` and `v0.1.0` is the owner's call.

- 2026-10-30: PR #8 CI confirmed terminal: all 4 check-runs success.
- 2026-10-30: Owner answered Q6 (Windows 11 only, ADR-0007) and
  deferred Q5 (video voice, ADR-0008).
- 2026-10-30: Step 9 on `feat/connectors-api`: connectors
  (manual, httpagent, claudecode with an injectable runner),
  session manager persisting events to SQLite, REST API v1 with
  bearer auth when a token is configured, CLI `session new`,
  `bait preview`, `demo` (leaky grades B/86, clean A/100), and
  `report` with `--explain-with` plus its disclosure line
  (ADR-0005). `spillcheck up` now serves dashboard + API + MCP on
  one port. Installers for Windows 11/macOS/Linux with checksum
  verification, distroless non-root Docker, goreleaser config and a
  SHA-pinned release workflow, and the bilingual README are in.
  Installer execution on real Windows/macOS machines is NOT verified
  here (config and syntax only). Local evidence: `go vet ./...`,
  `go test ./... -count=1` (21 packages ok), `go build`, demo runs
  in both locales.

- 2026-10-30: PR #7 CI confirmed terminal: all 4 check-runs success.
- 2026-10-30: Owner answered Q4/Q9; recorded in ADR-0005 (optional
  LLM report explanations with the user's own key, off by default,
  anonymised JSON only) and ADR-0006 (minimal brand: text wordmark,
  primary colour `#0E7C6B`).
- 2026-10-30: Step 8 on `feat/analyzer-reports`: causal analyzer
  with root-cause labels and good-behaviour detection
  (`internal/analyze`), transport/content-encoding assessment,
  localised summary, self-contained HTML and Markdown report
  renderers with masked bait values only, and the optional LLM
  explanation library (fake-client tested to prove no full bait
  value leaves the machine and the score never changes). Wiring
  `--explain-with` into the CLI with its disclosure text, full
  report sections (recommendations, per-bait table, limitations),
  and behaviour bonus scoring land with the session/report CLI in
  step 9; a code review of this PR fixed per-bait causal chains,
  evidence-based root causes, and scheme-accurate transport.
  Local evidence: `go vet ./...`, `go test ./... -count=1`
  (18 packages ok), `go build ./...`.

- 2026-10-30: PR #6 CI confirmed terminal: all 4 check-runs success.
- 2026-10-30: Step 7 on `feat/stage-injection-sinks`: fake blog with
  a hidden prompt-injection on post 3 (variants plain-hidden, base64,
  split, markdown-title), fake shop with hashed-card checkout,
  honeypot sinks (backup-sync, online-converter, payments,
  webhook-catcher; record and never forward), egress proxy recording
  `proxy.connect` with an SSRF guard for private destinations,
  scenarios T4-T8 plus the web injection definition, MCP
  `stage_fetch_url` serving the injection on post 3, and reports
  raising injection-driven sink sends to S4.
  Local evidence: `go vet ./...`, `go test ./... -count=1`
  (17 packages ok), `go build ./...`.

- 2026-10-30: PR #5 CI confirmed terminal: all 4 check-runs success.
- 2026-10-30: Step 6 on `feat/scenarios-detectors`: scenario YAML
  T1-T3 with a loader (`internal/scenario`, prompts VI/EN), basic
  detectors (`internal/detect`: fingerprint windows, digit
  normalisation, up to 4 decode layers, fragment joining, fictional
  facts at 2/3 components), first JSON report with scorer and
  coverage (`internal/report`, no full bait values), scripted demo
  agent profiles clean/leaky (`internal/demo`), and the first golden
  at `test/golden/report-basic.json`. Clean demo scores 100/A with
  0 findings; leaky is penalised. A scenario *runner* that drives a
  real agent through the YAML prompts is wired through the MCP task
  loop today; the YAML files drive prompts, the report builder
  evaluates `fail_if`-style exposures from events.
  Local evidence: `go vet ./...`, `go test ./... -count=1`
  (13 packages ok), `go build ./...`.

- 2026-10-30: Step 5 on `feat/mcp-server`: MCP server on the official
  Go SDK v1.2.0 (pinned because newer SDK releases require Go 1.25,
  PLAN minimum is 1.23), tools `spill_start`, `spill_next_task`,
  `spill_submit_result`, `spill_finish`, `spill_status` and all
  `stage_*` tools, prompt `start_spill_check`, stdio + Streamable
  HTTP transports (Bearer token enforced when configured), CLI
  `spillcheck mcp serve --stdio` and `spillcheck mcp config`.
  The backend holds one bait-seeded session in memory; scenario YAML,
  detectors and reports are step 6+. Tested end to end with the SDK's
  in-memory client transport. Local evidence: `go vet ./...`,
  `go test ./... -count=1` (11 packages ok), `go build ./...`.

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
