import type { AvailablePlan, BillingStatus, Plan } from "./types";

const checkoutPlanCodes = ["developer", "pro", "scale"];

export function availableCloudPlans(
  plans: AvailablePlan[],
  status: BillingStatus,
): AvailablePlan[] {
  if (
    status.deployment_mode !== "cloud" ||
    !status.billing_configured ||
    !status.checkout_available
  ) {
    return [];
  }

  return plans
    .filter(
      (plan) =>
        plan.status === "active" &&
        plan.pricing_type === "fixed" &&
        plan.amount_minor !== null &&
        plan.amount_minor > 0 &&
        checkoutPlanCodes.includes(plan.code),
    )
    .sort(
      (a, b) =>
        checkoutPlanCodes.indexOf(a.code) - checkoutPlanCodes.indexOf(b.code),
    );
}

export function canManageBilling(role: string): boolean {
  return role === "owner" || role === "admin";
}

export function formatCapacity(value: number | null): string {
  return value === null ? "Unlimited" : value.toLocaleString();
}

export function formatDate(value: string | null | undefined): string {
  if (!value) {
    return "Not available";
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "Not available";
  }

  return date.toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export function formatPrice(plan: Plan): string {
  if (plan.amount_minor === null) {
    return "Custom pricing";
  }
  if (plan.amount_minor === 0) {
    return "Free";
  }

  const price = new Intl.NumberFormat(undefined, {
    style: "currency",
    currency: plan.currency,
  }).format(plan.amount_minor / 100);

  return `${price} / ${plan.billing_interval}`;
}

export function hasConfirmedSubscription(status: BillingStatus): boolean {
  return (
    status.deployment_mode === "cloud" &&
    status.plan !== null &&
    status.plan.code !== "free" &&
    status.subscription?.provider === "polar" &&
    (status.subscription?.status === "active" ||
      status.subscription?.status === "trialing")
  );
}

export function runtimeUsagePercent(status: BillingStatus): number | null {
  if (status.usage.unlimited || status.usage.allowance_seconds <= 0) {
    return null;
  }

  return Math.min(
    100,
    Math.max(
      0,
      (status.usage.used_seconds / status.usage.allowance_seconds) * 100,
    ),
  );
}
