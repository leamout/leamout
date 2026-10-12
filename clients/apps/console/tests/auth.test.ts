import assert from "node:assert/strict";
import { test } from "node:test";
import { restoreFlow } from "../components/auth/auth-flow-provider";

const transactionId = "6f4772e1-2e28-4b1e-8d14-31c9a4a905a1";
const flow = {
  transactionId,
  email: "user@example.com",
  intent: "recovery",
  stage: "reset",
  expiresAt: 2000,
  resendAt: 1000,
};

test("restores only non-secret flow fields and rejects expired or corrupt state", () => {
  assert.deepEqual(
    restoreFlow(
      JSON.stringify({ ...flow, password: "secret", code: "012345" }),
      1000,
    ),
    flow,
  );
  assert.equal(restoreFlow(JSON.stringify(flow), 2000), null);
  assert.equal(restoreFlow("{broken", 1000), null);
  assert.equal(
    restoreFlow(JSON.stringify({ ...flow, transactionId: "invalid" }), 1000),
    null,
  );
  assert.equal(restoreFlow(null), null);
});
