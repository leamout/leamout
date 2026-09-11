---
name: roadmap-review
description: Review the current Leamout repository and recommend the next engineering work based on implemented capabilities, roadmap dependencies, risk, and product readiness. Use when asked what to build next.
---

# Review Leamout roadmap progress

Use the current repository as the source of truth. Read `AGENTS.md`, `README.md`, current roadmaps, recent merged PRs, open PRs, acceptance suites, and the relevant implementation directories.

## Evaluation

Assess:

- which roadmap capabilities are actually implemented and accepted;
- which documented items are stale or superseded;
- where a domain has schema/primitives but lacks an end-to-end product flow;
- production-readiness gaps versus synthetic acceptance coverage;
- financial, tenancy, security, carrier, and operational risks;
- whether proposed work unlocks a roadmap stage or merely adds optional complexity.

Prefer completing existing vertical slices over starting later roadmap stages. Do not recommend multi-carrier orchestration before single-carrier managed/BYOC reliability is proven. Do not recommend messaging/realtime/AI merely because scaffold packages exist.

Return a prioritized sequence of small PR-sized steps with the reason each one should precede the next.