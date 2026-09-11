---
name: review-pr
description: Review a Leamout pull request for correctness, architecture boundaries, regressions, security, tenancy, idempotency, generated artifacts, and acceptance coverage. Use when asked to review a PR or code change.
---

# Review a Leamout pull request

Read `AGENTS.md` first, then inspect the PR metadata, changed files, diff, relevant current code, tests, and documentation.

## Review order

1. Identify the user-visible or architectural behavior the PR intends to change.
2. Check dependency direction and domain ownership before style details.
3. Check tenancy, authorization, idempotency, transaction boundaries, retries, and failure behavior.
4. For Commercial changes, verify monetary invariants, provider/adaptor boundaries, immutable ledger behavior, and SQLC-only persistence.
5. For Telecom changes, verify BYOC/managed isolation, no silent fallback, routing ownership, and adapter boundaries.
6. Check generated artifacts are reproducible from source definitions.
7. Check tests prove the intended behavior without weakening existing acceptance guarantees.
8. Inspect CI failures and distinguish implementation failures from infrastructure flakiness.

## Output

Lead with concrete findings ordered by severity. Cite exact files/functions/lines when available. Do not invent problems merely to populate a review. If no blocking issue is found, say so and list any residual risks or missing coverage separately.

Do not modify the branch unless the user explicitly asks for fixes.