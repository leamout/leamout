import assert from "node:assert/strict";
import { test } from "node:test";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { BillingOverview } from "../components/billing/billing-overview";
import {
  availableCloudPlans,
  canManageBilling,
  hasConfirmedSubscription,
  runtimeUsagePercent,
} from "../components/billing/presentation";
import type { AvailablePlan, BillingStatus } from "../components/billing/types";

function plan(
  code: string,
  overrides: Partial<AvailablePlan> = {},
): AvailablePlan {
  return {
    code,
    name: code,
    currency: "USD",
    amount_minor: 7900,
    billing_interval: "month",
    pricing_type: "fixed",
    status: "active",
    limits: {},
    ...overrides,
  };
}

function billingStatus(overrides: Partial<BillingStatus> = {}): BillingStatus {
  return {
    deployment_mode: "cloud",
    billing_configured: true,
    checkout_available: true,
    portal_available: true,
    plan: plan("free", {
      amount_minor: 0,
    }),
    subscription: null,
    limits: {
      max_concurrent_calls: 1,
      retention_days: 7,
      monthly_runtime_minutes: 100,
      max_voice_agents: 1,
      max_sip_trunks: 1,
      period_start: "2026-10-01T00:00:00Z",
      period_end: "2026-11-01T00:00:00Z",
    },
    usage: {
      used_seconds: 1200,
      allowance_seconds: 6000,
      unlimited: false,
      period_start: "2026-10-01T00:00:00Z",
      period_end: "2026-11-01T00:00:00Z",
    },
    ...overrides,
  };
}

test("only supported active paid Cloud plans can be purchased", () => {
  const plans = [
    plan("scale"),
    plan("free", {
      amount_minor: 0,
    }),
    plan("developer"),
    plan("pro"),
    plan("enterprise"),
    plan("developer", {
      status: "archived",
    }),
    plan("pro", {
      pricing_type: "custom",
    }),
    plan("scale", {
      amount_minor: null,
    }),
  ];

  assert.deepEqual(
    availableCloudPlans(plans, billingStatus()).map((item) => item.code),
    ["developer", "pro", "scale"],
  );
  assert.equal(plans[0].code, "scale");
});

test("self-hosting and unavailable billing never offer Cloud purchases", () => {
  const plans = [plan("developer")];
  const unavailableStates: Partial<BillingStatus>[] = [
    {
      deployment_mode: "self_hosted",
    },
    {
      billing_configured: false,
    },
    {
      checkout_available: false,
    },
  ];

  for (const state of unavailableStates) {
    assert.deepEqual(availableCloudPlans(plans, billingStatus(state)), []);
  }
});

test("only organization owners and administrators manage billing", () => {
  assert.equal(canManageBilling("owner"), true);
  assert.equal(canManageBilling("admin"), true);

  for (const role of ["member", "viewer", "", "OWNER"]) {
    assert.equal(canManageBilling(role), false);
  }
});

test("a paid plan selection does not confirm a checkout without an active subscription", () => {
  const status = billingStatus({
    plan: plan("pro"),
  });

  assert.equal(hasConfirmedSubscription(status), false);

  assert.equal(
    hasConfirmedSubscription({
      ...status,
      subscription: {
        provider: null,
        status: "active",
        current_period_end: null,
        cancel_at_period_end: false,
        past_due_at: null,
        grace_ends_at: null,
      },
    }),
    false,
  );

  for (const subscriptionStatus of [
    "incomplete",
    "past_due",
    "canceled",
    "unpaid",
  ]) {
    assert.equal(
      hasConfirmedSubscription({
        ...status,
        subscription: {
          provider: "polar",
          status: subscriptionStatus,
          current_period_end: "2026-11-01T00:00:00Z",
          cancel_at_period_end: false,
          past_due_at: null,
          grace_ends_at: null,
        },
      }),
      false,
    );
  }

  for (const subscriptionStatus of ["active", "trialing"]) {
    const confirmed = billingStatus({
      plan: plan("pro"),
      subscription: {
        provider: "polar",
        status: subscriptionStatus,
        current_period_end: "2026-11-01T00:00:00Z",
        cancel_at_period_end: true,
        past_due_at: null,
        grace_ends_at: null,
      },
    });

    assert.equal(hasConfirmedSubscription(confirmed), true);
    assert.equal(
      hasConfirmedSubscription({
        ...confirmed,
        deployment_mode: "self_hosted",
      }),
      false,
    );
    assert.equal(
      hasConfirmedSubscription({
        ...confirmed,
        plan: plan("free"),
      }),
      false,
    );
  }
});

test("unlimited self-hosted usage does not display a Cloud quota", () => {
  const status = billingStatus();

  assert.equal(runtimeUsagePercent(status), 20);
  assert.equal(
    runtimeUsagePercent({
      ...status,
      usage: {
        ...status.usage,
        used_seconds: 10000,
      },
    }),
    100,
  );
  assert.equal(
    runtimeUsagePercent({
      ...status,
      deployment_mode: "self_hosted",
      usage: {
        ...status.usage,
        unlimited: true,
      },
    }),
    null,
  );
});

test("overdue Cloud billing displays whether paid grace limits remain effective", () => {
  const status = billingStatus({
    plan: plan("pro"),
    subscription: {
      provider: "polar",
      status: "past_due",
      current_period_end: "2026-11-01T00:00:00Z",
      cancel_at_period_end: false,
      past_due_at: "2026-10-13T00:00:00Z",
      grace_ends_at: new Date(Date.now() + 86400000).toISOString(),
    },
  });

  const graceMarkup = renderToStaticMarkup(
    createElement(BillingOverview, {
      status,
    }),
  );

  assert.match(graceMarkup, /Paid limits remain available until/);
  assert.doesNotMatch(
    graceMarkup,
    /Free limits apply until payment is recovered/,
  );

  const expiredMarkup = renderToStaticMarkup(
    createElement(BillingOverview, {
      status: {
        ...status,
        subscription: {
          ...status.subscription,
          provider: "polar",
          status: "past_due",
          current_period_end: "2026-11-01T00:00:00Z",
          cancel_at_period_end: false,
          past_due_at: "2026-10-13T00:00:00Z",
          grace_ends_at: new Date(Date.now() - 86400000).toISOString(),
        },
      },
    }),
  );

  assert.match(expiredMarkup, /Free limits apply until payment is recovered/);
  assert.doesNotMatch(expiredMarkup, /Paid limits remain available until/);
});

test("self-hosted overview shows licensed capacity without Cloud billing or a usage quota", () => {
  const status = billingStatus({
    deployment_mode: "self_hosted",
    plan: null,
    subscription: null,
    billing_configured: false,
    checkout_available: false,
    portal_available: false,
    license: {
      state: "valid",
      license: {
        license_id: "license-production",
        customer_id: "customer-first",
        deployment_id: "deployment-production",
        valid_from: "2026-10-01T00:00:00Z",
        expires_at: "2027-10-01T00:00:00Z",
        grace_until: "2027-10-15T00:00:00Z",
        max_concurrent_calls: 50,
        max_nodes: 3,
        features: ["sso", "audit_logging"],
      },
    },
  });
  status.limits.monthly_runtime_minutes = null;
  status.limits.max_concurrent_calls = 50;
  status.usage.unlimited = true;

  const markup = renderToStaticMarkup(
    createElement(BillingOverview, {
      status,
    }),
  );

  assert.match(markup, /Self-hosted license/);
  assert.match(markup, /license-production/);
  assert.match(markup, /Licensed nodes/);
  assert.match(markup, /Licensed concurrent calls/);
  assert.match(markup, /audit logging/);
  assert.match(markup, /No monthly runtime minute quota/);
  assert.doesNotMatch(markup, /Current effective plan/);
  assert.doesNotMatch(markup, /Purchases and payment updates/);
  assert.doesNotMatch(markup, /<progress/);
});
