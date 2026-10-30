# Progress

Updated at the end of every working session (PLAN 0.1.8).

## Done

- 2026-10-30: Read `docs/PLAN.md` in full. Owner answered Q1 and Q3.
- 2026-10-30: ADR-0001 (GitHub owner `Hungleminh`), ADR-0002 (ports 7700-7709).
- 2026-10-30: Step 1 of checklist 24.6 in progress on `feat/bootstrap-repo`:
  `docs/PLAN.md`, `NOTICE`, `SECURITY.md`, `AGENTS.md`, `.gitleaks.toml`,
  hardened `.gitignore`, Dependabot config.

## In progress

- Step 1: CodeQL workflow file; GitHub-side settings still manual (below).
- Step 2 of checklist 24.6: Go module, cobra CLI (`version`, `doctor`,
  `up` with an empty dashboard), `ci.yml`.

## Waiting on the project owner

- GitHub repo settings that cannot be set from code: enable secret
  scanning + push protection, and branch protection on `main`
  (Settings -> Code security / Branches). No `gh` CLI in this environment.
- Q2, Q4-Q10 remain open; each will be asked when its step is reached
  (Q7/Q8 before bait/persona, Q4/Q9 before HTML report, Q6 before
  installers, Q5 before video, Q2 before v0.2, Q10 at CI hardening).
