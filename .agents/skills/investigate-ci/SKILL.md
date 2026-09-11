---
name: investigate-ci
description: Diagnose and fix failing Leamout CI or acceptance workflows. Use when a GitHub Actions job, acceptance gate, test, lint, generation, or build check fails.
---

# Investigate Leamout CI

Read `AGENTS.md` and the relevant acceptance or subsystem documentation before changing code.

1. Inspect the exact failed workflow, job, step, and log output.
2. Find the first meaningful failure, not merely the final exit code.
3. Correlate it with the changed code and current runtime topology.
4. Reproduce locally when practical using the same command CI runs.
5. Fix the implementation or environment contract. Do not weaken assertions, skip readiness checks, remove coverage, or replace real telecom behavior with mocks just to make CI pass.
6. Run the narrowest relevant verification first, then broader checks required by the repository workflow.
7. Re-read logs after the fix and confirm no secondary failure is hidden behind the first one.

For telecom failures, inspect startup/readiness, OpenSIPS, RTPengine, FreeSWITCH, Redis, NATS, provider adapters, and network boundaries as relevant. For generated-code failures, change the source definition and regenerate rather than hand-editing generated output.

Report the root cause, the fix, and the checks that prove it.