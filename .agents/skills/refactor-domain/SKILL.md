---
name: refactor-domain
description: Refactor a Leamout domain or package boundary without changing product behavior. Use when reorganizing modules, services, repositories, adapters, files, or dependency direction.
---

# Refactor a Leamout domain

Read `AGENTS.md`, the relevant domain docs, composition root, public routes, repositories, tests, and acceptance coverage before editing.

## Method

1. State the current ownership problem in terms of responsibilities and dependency direction.
2. Define the target boundary before moving files.
3. Preserve public behavior unless the task explicitly changes it.
4. Prefer one clear application service boundary over callers reaching into repositories, registries, or provider implementations separately.
5. Keep external provider logic behind adapters and durable persistence behind repositories.
6. Avoid compatibility aliases unless a real migration requires them; remove transient bridges once callers migrate.
7. Keep file names responsibility-based (`model.go`, `service.go`, `repository.go`, `validation.go`, `handler.go`, `routes.go`, `consumer.go`, `publisher.go`, `jobs.go`) and do not add empty scaffolds.
8. Update composition and tests in the same change.
9. Run formatting, focused tests, broader required checks, and generation checks.

Do not mix unrelated feature work into an architectural refactor.