# Review Roadmap and Recommend Next Work

Review the current Leamout repository and recommend the next engineering work based on the implementation that exists now, not only the roadmap text.

Use `AGENTS.md` and current documents under `docs/` as context, but verify claims against current code, tests, and recent merged work.

Assess:

- which roadmap capabilities are genuinely complete;
- which foundations exist but are not connected into end-to-end product behavior;
- the highest-risk missing vertical slice;
- stale documentation that can mislead future work;
- whether proposed work belongs to the current roadmap stage or is premature.

Prioritize work that:

1. closes a real product/invariant gap;
2. protects money, tenant isolation, telecom reliability, or provider correctness;
3. creates an end-to-end acceptance guarantee;
4. enables the next roadmap stage without introducing speculative architecture.

Do not recommend messaging, multi-carrier/LCR, AI media, or other later-stage features merely because their directories or docs exist.

Return:

- current state summary;
- top three next initiatives in priority order;
- the exact first PR you would create;
- acceptance criteria for that PR;
- work that should explicitly wait.
