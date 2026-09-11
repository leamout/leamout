# Refactor Domain Boundary

Refactor the requested Leamout domain/package without changing product behavior unless the task explicitly asks for a contract change.

Before editing:

- read `AGENTS.md`;
- read the relevant `.github/instructions/*.instructions.md` file;
- inspect the current package, callers, composition root, persistence queries, and tests;
- identify which responsibilities are leaking across the boundary.

Refactor toward:

- narrow service interfaces;
- provider-neutral core models;
- repositories limited to persistence;
- services owning business rules/orchestration;
- handlers limited to HTTP transport;
- SQLC-only Commercial persistence;
- no empty scaffold files;
- no compatibility aliases unless a staged migration truly needs them.

Preserve idempotency, tenant ownership, public API behavior, and acceptance guarantees.

Prefer one coherent atomic change over a sequence of partial compatibility layers.

After editing, verify formatting, tests, linting, generators, and relevant acceptance gates.

Summarize the final dependency direction and explain why the new boundary is better than the previous one.
