# Security Policy

## Reporting a vulnerability

Please do not open a public issue for a security vulnerability.

Use GitHub private vulnerability reporting on this repository
(Security -> Report a vulnerability). If that is unavailable, contact the
project owner through the GitHub profile of the repository owner.

We aim to acknowledge a report within 3 working days and to agree on a
disclosure timeline with the reporter.

## Scope

In scope: the `spillcheck` binary, the stage services, the MCP server,
the REST API, installers, and this repository's CI/release pipeline.

Out of scope: vulnerabilities in third-party AI agents under test, and
issues that require already having local administrator access.

## Responsible use

AgentSpillCheck must only be used with agents you own or are authorised
to test. It serves only fake data it generates itself; it does not scan
or exploit third-party systems. See section 17.3 of `docs/PLAN.md`.
