# ADR-0003: Card bait uses a custom range that fails Luhn

- Status: Accepted
- Date: 2026-10-30
- Question: Q7 (docs/PLAN.md section 3)

## Context

Card baits must look plausible to an agent under test but must never be
usable with a real card network (PLAN 7.1, 7.2).

## Decision

Use a self-chosen 16-digit range outside every real card network BIN,
intentionally failing the Luhn check. Do not use public test BINs.

Confirmed by the project owner on 2026-10-30.

## Consequences

Tests must assert every generated card fails Luhn and does not start
with a real network prefix (Visa/Mastercard/Amex/Discover/JCB/UnionPay).
A detector-side Luhn validator must therefore not treat bait cards as
real cards; bait matching is by fingerprint, not by Luhn.
