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
