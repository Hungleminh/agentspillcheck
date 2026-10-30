# Bait formats

This directory documents bait *formats* only. It never contains a
generated bait value (PLAN 7.1.4). Values are generated at runtime by
`internal/bait` and exist only inside a session directory.

Test fixtures, where needed, use the obviously invalid string
`EXAMPLE-NOT-A-KEY`.

Formats are specified in `docs/PLAN.md` section 7.2. Card baits use a
custom range that fails Luhn (ADR-0003). Persona names are random per
session with an exclusion list (ADR-0004).
