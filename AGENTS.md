# AGENTS.md

Guidance for coding agents working in this repository.

## Product in one sentence

AgentSpillCheck feeds fake sensitive data to an AI agent and reports what
it read, where it sent it, where it kept it, and whether it was encrypted,
without ever touching the user's real data.

## Governing document

The product spec and ADR log are maintained locally by the owner and are
not published in this repo. For any open product question: stop and ask
the project owner, and never assume an answer.

## Layout

Code lives under
`cmd/spillcheck/` and `internal/`; public schemas under `pkg/schema/`.
Ready-made HTML report samples live under `report/samples/`; the
`report/generate-report/SKILL.md` skill teaches agents to build that
HTML from a `report.md`.

## Setup

Prerequisites: Go 1.26+ and Docker (for container runs).

```sh
go build ./...        # or: make build
go test ./...         # or: make test (add -count=1 to skip cache)
go vet ./...          # or: make vet
```

Install paths: `installers/install.sh` (macOS/Linux),
`installers/install.bat` (Windows 11, the only supported Windows
version), or Docker via `deploy/docker-compose.yml`.

## Everyday workflows

Try a scripted leaky agent and print its grade:

```sh
spillcheck demo --profile leaky --locale en
spillcheck demo --profile clean --locale vi --out ./out
```

`--out` writes `report.json`, `report.html` and `report.md` for the
session. Serve the dashboard and API (binds loopback by default,
ports 7700–7709):

```sh
spillcheck up
```

Connect an agent over MCP (Claude Desktop, Cursor, Claude Code):

```sh
spillcheck mcp config --client claude-desktop
spillcheck mcp serve --stdio
```

The agent flow is `spill_start` → `spill_next_task` → `stage_*`
tools → `spill_submit_result` → `spill_finish`. Scenario files live
under `scenarios/T1.yaml` … `T8.yaml`.

## Generating the HTML report

`report.html` from `--out` is already the styled single-file report.
To rebuild that style from any `report.md` by hand (or to restyle
it), load the skill at `report/generate-report/SKILL.md` first: it
defines the input map, the CSS token system with dark mode, the
layout order, and the acceptance checklist. Reference renders are in
`report/samples/`. Rules that always apply: one file, inline CSS
only, no JavaScript, no external assets, masked bait samples only,
numbers copied exactly.

## Commands

```sh
make build   # go build ./...
make test    # go test ./...
make vet     # go vet ./...
```

## Rules

Forbidden:

- Committing any real bait value, secret, token, `.env`, session data,
  or real personal data. The repo contains formats and dictionaries only.
- Adding telemetry or any "phone home" behaviour.
- Calling external services on the detection path, or using an LLM in
  detection (no LLM is used in detection).
- Hardcoding user-visible strings. All display strings go through i18n
  VI/EN (`internal/core/i18n`).

Required:

- Apache-2.0 licence header in every source file.
- Update both VI and EN i18n locales for every new display string.
- Update schema and golden files when the report format changes.
- Run gitleaks before committing when it is available.
- Small commits, Conventional Commits, English messages.
- Work on `feat/...` branches; open a PR per checklist step; never push
  directly to `main`.
- Bind `127.0.0.1` by default; never log full bait values at INFO.
- Progress tracking is local-only (owner decision): never commit or push progress notes.

## Security

This repo is public from the first commit: never commit real bait
values, secrets, tokens, `.env` files, session data, or real personal
data; keep telemetry out; bind `127.0.0.1` by default; never log full
bait values at INFO.
