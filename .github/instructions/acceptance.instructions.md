---
applyTo: "tests/acceptance/**/*,.github/workflows/*.yml,.github/workflows/*.yaml"
---

# Acceptance and CI Instructions

Follow `AGENTS.md` first.

Acceptance tests are executable product contracts, not convenience tests.

## Rules

- Do not remove assertions, skip scenarios, replace real telecom behavior with mocks, or weaken readiness merely to make CI pass.
- Fix the implementation when a product guarantee regresses.
- Keep each acceptance topology isolated and explicit about the delivery mode it proves.
- Preserve real boundaries where the suite is intended to exercise them: SIP signaling, media relay, FreeSWITCH control, TURN/WebRTC, managed routing, provider reconciliation, database state, and restart behavior.
- Prefer deterministic readiness/health checks over arbitrary sleep increases.
- Do not hide flaky startup ordering by retrying forever; define bounded readiness and useful diagnostics.
- When a workflow fails, identify the first causal error rather than patching downstream symptoms.
- Keep generated-artifact checks strict: a generator diff is a source-of-truth problem, not something to ignore.

## Delivery-mode isolation

- BYOC tests must not silently use managed connectivity.
- Managed tests must not consume tenant BYOC resources.
- Cloud and self-hosted tests must preserve their intended runtime-placement boundary.
- Synthetic provider acceptance does not prove production carrier readiness; keep production-provider evidence separate.

## CI changes

Workflow changes should improve signal, diagnostics, reproducibility, or correctness. Do not disable lint, vulnerability scanning, generation checks, or acceptance gates without an explicit repository decision and documented reason.
