# GitHub Copilot Repository Instructions

Follow the repository engineering contract in `AGENTS.md`.

Apply these rules to all Copilot work in this repository:

- Read the current implementation and relevant `docs/` before changing architecture.
- Preserve BYOC as a first-class path and never introduce silent BYOC/managed fallback.
- Keep provider-specific logic behind adapters in `server/internal/integrations`.
- Preserve tenant isolation and idempotency at durable boundaries.
- Commercial uses subscription/license plus prepaid collection; do not introduce postpaid usage credit without an explicit product decision.
- Payments reconciles money; Checkout interprets successful settlement; Wallets own prepaid value.
- PostgreSQL is authoritative durable/monetary state. Redis must not create or settle value.
- Do not weaken acceptance tests to make CI pass.
- Do not edit generated SQLC or Templ Go output by hand when the source definition should be changed instead.
- Keep PRs focused and update docs when architecture or public behavior changes.
- Before finishing Go changes, run the repository's current formatting, tests, linting, and required generators as defined by CI/Makefile.
