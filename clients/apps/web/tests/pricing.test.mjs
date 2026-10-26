import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";
import { plans } from "../lib/plans.ts";

const sql = readFileSync(
  new URL("../../../../server/migrations/033_create_plans.sql", import.meta.url),
  "utf8",
);
const values = sql.match(
  /INSERT INTO plans \([\s\S]*?\)\s*VALUES\s*([\s\S]*?)\s*ON CONFLICT \(code\)/,
);
assert.ok(values, "Expected an INSERT INTO plans seed");

const rowPattern =
  /\(\s*'(?<code>[^']+)',\s*'(?<name>[^']+)',\s*'(?:[^']|'')*',\s*'(?<pricingType>fixed|custom)',\s*'(?<currency>[A-Z]{3})',\s*(?<amountMinor>\d+|NULL),\s*'(?<billingInterval>month|year)',\s*'(?<entitlements>\{[^']*\})'::jsonb,\s*'(?<limits>\{[^']*\})'::jsonb,\s*'(?<status>active|inactive)'\s*\)/g;

const rows = Array.from(values[1].matchAll(rowPattern), ({ groups }) => ({
  code: groups.code,
  name: groups.name,
  pricingType: groups.pricingType,
  currency: groups.currency,
  amountMinor: groups.amountMinor,
  billingInterval: groups.billingInterval,
  status: groups.status,
  entitlements: JSON.parse(groups.entitlements),
  limits: JSON.parse(groups.limits),
}));

test("all SQL seed rows are parsed", () => {
  const codes = Array.from(
    values[1].matchAll(/^\s*\(\s*'([a-z0-9_-]+)'/gm),
    (match) => match[1],
  );
  assert.ok(codes.length > 0);
  assert.deepEqual(rows.map(({ code }) => code), codes);
});

test("public monthly pricing matches the backend seed", () => {
  const publicRows = rows.filter(
    (row) => row.pricingType === "fixed" && row.status === "active",
  );
  assert.deepEqual(
    publicRows.map(({ code }) => code).sort(),
    plans.map(({ code }) => code).sort(),
  );

  for (const plan of plans) {
    const row = publicRows.find(({ code }) => code === plan.code);
    assert.ok(row, `No seeded plan for ${plan.code}`);
    assert.equal(row.name, plan.name);
    assert.equal(row.currency, "USD");
    assert.equal(row.billingInterval, "month");
    assert.equal(Number(row.amountMinor), plan.price * 100);
    assert.equal(row.limits.monthly_runtime_minutes, plan.minutes);
    assert.equal(row.limits.max_concurrent_calls, plan.concurrency);
    assert.equal(row.limits.max_voice_agents, plan.agents);
    assert.equal(row.limits.max_sip_trunks, plan.trunks);
    assert.equal(row.limits.retention_days, plan.retention);

    const hasAdvancedControls = plan.code === "pro" || plan.code === "scale";
    assert.equal(row.entitlements.advanced_rbac, hasAdvancedControls);
    assert.equal(row.entitlements.retention_policies, hasAdvancedControls);
  }
});

test("custom plans are not presented as monthly fixed-price cards", () => {
  for (const row of rows.filter(({ pricingType }) => pricingType === "custom")) {
    assert.ok(!plans.some(({ code }) => code === row.code));
  }
});
