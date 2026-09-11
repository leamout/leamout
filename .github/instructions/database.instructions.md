---
applyTo: "server/internal/database/queries/*.sql,server/internal/database/sqlc/*.go,server/migrations/*.sql,server/migrations/atlas.sum"
---

# Database and Generation Instructions

Follow `AGENTS.md` first.

## SQLC source of truth

- Application SQL belongs in `server/internal/database/queries/*.sql`.
- Generated bindings live in `server/internal/database/sqlc`.
- Change query sources first; regenerate SQLC; do not hand-fix generated bindings.
- Keep generated output reproducible with the repository-pinned/current generator command used by CI.

## Query design

- Enforce organization/resource ownership in SQL predicates and constraints where practical.
- Use deterministic ordering for list/query results when API behavior depends on order.
- Make idempotent operations explicit with suitable unique constraints and conflict behavior.
- Use row locking/transactions for monetary admission and other concurrent state transitions that require serialization.
- Prefer typed SQL expressions/casts that allow SQLC to generate concrete Go types instead of `interface{}`-like ambiguity.

## Migrations

- Treat applied/released migrations as immutable.
- Only edit existing migration history when repository documentation explicitly says the series is still pre-release and the active task intends consolidation.
- Keep `atlas.sum` consistent with migration sources.
- New constraints should encode durable invariants rather than relying only on application validation.

## Financial data

- Monetary values use integer representations appropriate to the domain; do not introduce floating-point monetary storage.
- Preserve immutable ledger/provider-event/history records.
- Keep customer prices separate from upstream wholesale cost.

## Verification

When query or migration sources change:

1. regenerate SQLC;
2. regenerate/update Atlas checksums when required;
3. run targeted tests;
4. run the Server generation/clean-diff checks used by CI;
5. ensure generated files contain no unexplained manual edits.
