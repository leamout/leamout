# Investigate CI Failure

Investigate the current failing CI run using Leamout's repository instructions.

Work from the first causal failure, not the final exit code or downstream symptoms.

Process:

1. identify the earliest failing job/step and extract the exact error;
2. inspect the implementation and recent changes that can cause it;
3. distinguish source-of-truth failures from generated-artifact failures;
4. distinguish startup/readiness failures from real product-behavior failures;
5. propose the smallest fix that preserves existing product contracts;
6. do not weaken tests, skip checks, increase arbitrary sleeps, or disable lint/generation/security gates merely to make CI pass;
7. run or recommend the targeted tests that prove the root cause is fixed.

For telecom failures, preserve BYOC/managed isolation and real signaling/media behavior.

For Commercial failures, preserve tenant ownership, payment/wallet idempotency, immutable money history, and prepaid authorization.

Finish with:

- root cause;
- files to change;
- exact fix;
- verification commands/gates;
- any remaining risk.
