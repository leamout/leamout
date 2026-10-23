"use client";

import { Button } from "@leamout/ui/components/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@leamout/ui/components/card";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { ApiError, api, errorMessage } from "@/lib/api";
import { BillingOverview } from "./billing-overview";
import {
  availableCloudPlans,
  canManageBilling,
  formatCapacity,
  formatPrice,
  hasConfirmedSubscription,
} from "./presentation";
import type { AvailablePlan, BillingStatus, Organization } from "./types";

const checkoutPollAttempts = 10;
const checkoutPollInterval = 3_000;

export function BillingSettings() {
  const router = useRouter();
  const params = useSearchParams();
  const requestedOrganization = params.get("organization");
  const [busy, setBusy] = useState(false);
  const organizationContext = useOrganization();
  const activeOrganizations = organizationContext.organizations.filter(
    (organization) => organization.status === "active",
  );
  const selected = requestedOrganization
    ? activeOrganizations.find(
        (organization) => organization.id === requestedOrganization,
      )
    : organizationContext.selected;

  return (
    <div className="space-y-8">
      <header className="space-y-2">
        <h1 className="text-2xl font-semibold">Billing</h1>
        <p className="text-muted-foreground">
          Review your plan, usage, and payment settings.
        </p>
      </header>
      {organizationContext.isPending || organizationContext.isFetching ? (
        <p role="status" className="text-sm text-muted-foreground">
          Loading your organizations…
        </p>
      ) : organizationContext.isError ? (
        <RequestError
          error={organizationContext.error}
          onRetry={organizationContext.refetch}
        />
      ) : activeOrganizations.length ? (
        <>
          <div className="max-w-sm space-y-2">
            <label
              htmlFor="billing-organization"
              className="text-sm font-medium"
            >
              Organization
            </label>
            <select
              id="billing-organization"
              className="h-10 w-full rounded-lg border bg-background px-3 text-sm"
              value={selected?.id ?? ""}
              disabled={busy}
              onChange={(event) => {
                if (
                  organizationContext.selectOrganization(event.target.value)
                ) {
                  const query = new URLSearchParams({
                    organization: event.target.value,
                  });
                  router.replace(`/settings/billing?${query}`);
                }
              }}
            >
              {!selected && <option value="">Choose an organization</option>}
              {activeOrganizations.map((organization) => (
                <option key={organization.id} value={organization.id}>
                  {organization.name}
                </option>
              ))}
            </select>
          </div>
          {selected ? (
            <BillingAccount
              key={selected.id}
              organization={selected}
              checkoutResult={params.get("checkout")}
              onBusyChange={setBusy}
            />
          ) : (
            <p role="alert" className="text-sm text-muted-foreground">
              This organization is unavailable. Select an organization you can
              access.
            </p>
          )}
        </>
      ) : (
        <div className="space-y-4 rounded-lg border border-dashed p-6">
          <p>Create an organization before managing billing.</p>
          <Button render={<Link href="/create-organization" />}>
            Create organization
          </Button>
        </div>
      )}
    </div>
  );
}

function BillingAccount({
  organization,
  checkoutResult,
  onBusyChange,
}: {
  organization: Organization;
  checkoutResult: string | null;
  onBusyChange: (busy: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const canManage = canManageBilling(organization.role);
  const [polls, setPolls] = useState(0);
  const checkoutKeys = useRef<Record<string, string>>({});
  const status = useQuery({
    queryKey: ["billing-status", organization.id],
    queryFn: ({ signal }) =>
      api<BillingStatus>("/billing/status", {
        signal,
        organizationId: organization.id,
      }),
    staleTime: 0,
    retry: false,
  });
  const plans = useQuery({
    queryKey: ["billing-plans"],
    queryFn: ({ signal }) =>
      api<{ plans: AvailablePlan[] }>("/plans", { signal }),
    enabled:
      status.data?.deployment_mode === "cloud" &&
      status.data.billing_configured &&
      status.data.checkout_available &&
      canManage,
  });

  const checkout = useMutation({
    mutationFn: async (planCode: string) => {
      const returnURL = billingReturnURL(organization.id);
      const successURL = new URL(returnURL);
      successURL.searchParams.set("checkout", "success");
      const cancelURL = new URL(returnURL);
      cancelURL.searchParams.set("checkout", "canceled");
      const key = checkoutKeys.current[planCode] ?? crypto.randomUUID();
      checkoutKeys.current[planCode] = key;

      const result = await api<{ url: string }>("/billing/checkout", {
        method: "POST",
        organizationId: organization.id,
        idempotencyKey: key,
        body: {
          plan_code: planCode,
          success_url: successURL.href,
          cancel_url: cancelURL.href,
        },
      });
      redirectToBilling(result.url);
    },
    onError: () => {
      status.refetch();
    },
  });
  const portal = useMutation({
    mutationFn: async () => {
      const result = await api<{ url: string }>("/billing/portal", {
        method: "POST",
        organizationId: organization.id,
        body: {
          return_url: billingReturnURL(organization.id),
        },
      });
      redirectToBilling(result.url);
    },
  });

  const confirmed = status.data ? hasConfirmedSubscription(status.data) : false;
  const waitingForCheckout =
    canManage &&
    checkoutResult === "success" &&
    status.data?.deployment_mode === "cloud" &&
    !confirmed;

  useEffect(() => {
    onBusyChange(checkout.isPending || portal.isPending);
    return () => onBusyChange(false);
  }, [checkout.isPending, portal.isPending, onBusyChange]);

  useEffect(() => {
    if (
      !waitingForCheckout ||
      status.isError ||
      status.isFetching ||
      polls >= checkoutPollAttempts
    ) {
      return;
    }

    const timer = window.setTimeout(() => {
      setPolls((attempts) => attempts + 1);
      api<BillingStatus>("/billing/reconcile", {
        method: "POST",
        organizationId: organization.id,
      })
        .then((result) =>
          queryClient.setQueryData(["billing-status", organization.id], result),
        )
        .catch(() => status.refetch());
    }, checkoutPollInterval);

    return () => window.clearTimeout(timer);
  }, [
    waitingForCheckout,
    polls,
    status.isError,
    status.isFetching,
    status.refetch,
    queryClient,
    organization.id,
  ]);

  if (status.isPending) {
    return (
      <p role="status" className="text-sm text-muted-foreground">
        Loading your billing details…
      </p>
    );
  }
  if (status.isError) {
    return (
      <RequestError error={status.error} onRetry={() => status.refetch()} />
    );
  }

  const value = status.data;
  const eligiblePlans = availableCloudPlans(plans.data?.plans ?? [], value);
  const actionError = checkout.error ?? portal.error;
  const pending = checkout.isPending || portal.isPending;

  return (
    <div className="space-y-6">
      {checkoutResult === "success" && value.deployment_mode === "cloud" && (
        <p role="status" className="rounded-lg border p-4 text-sm">
          {confirmed
            ? "Your subscription is confirmed. Your current plan and limits are shown below."
            : polls < checkoutPollAttempts
              ? "Checkout returned. Waiting for your confirmed subscription…"
              : "Your subscription has not been confirmed yet. Refresh the status shortly. Contact support if payment was completed and your plan remains unchanged."}
        </p>
      )}
      {checkoutResult === "canceled" && value.deployment_mode === "cloud" && (
        <p role="status" className="text-sm text-muted-foreground">
          Checkout was canceled. Your current plan remains in effect.
        </p>
      )}
      <BillingOverview status={value} />
      <div className="flex flex-wrap gap-3">
        {canManage &&
          value.deployment_mode === "cloud" &&
          value.portal_available && (
            <Button disabled={pending} onClick={() => portal.mutate()}>
              {portal.isPending ? "Opening portal…" : "Manage subscription"}
            </Button>
          )}
        <Button
          variant="outline"
          disabled={status.isFetching || pending}
          onClick={() => status.refetch()}
        >
          {status.isFetching ? "Refreshing…" : "Refresh status"}
        </Button>
      </div>
      {value.deployment_mode === "cloud" && value.portal_available && (
        <p className="text-sm text-muted-foreground">
          {canManage
            ? "Use the customer portal to update payment details, change plans, or cancel your subscription."
            : "An organization owner or administrator can manage payment details and subscriptions."}
        </p>
      )}
      {actionError && (
        <p role="alert" className="text-sm text-destructive">
          {billingErrorMessage(actionError)}
        </p>
      )}
      {canManage &&
        value.deployment_mode === "cloud" &&
        value.checkout_available && (
          <section className="space-y-4" aria-labelledby="cloud-plans-title">
            <div className="space-y-2">
              <h2 id="cloud-plans-title" className="text-lg font-semibold">
                Cloud plans
              </h2>
              <p className="text-sm text-muted-foreground">
                Choose a plan for this organization. Purchases are completed
                securely with Polar.
              </p>
            </div>
            {plans.isPending ? (
              <p role="status" className="text-sm text-muted-foreground">
                Loading available plans…
              </p>
            ) : plans.isError ? (
              <RequestError
                error={plans.error}
                onRetry={() => plans.refetch()}
              />
            ) : eligiblePlans.length ? (
              <div className="grid gap-4 lg:grid-cols-3">
                {eligiblePlans.map((plan) => (
                  <PlanCard
                    key={plan.code}
                    plan={plan}
                    disabled={pending || waitingForCheckout}
                    pending={
                      checkout.isPending && checkout.variables === plan.code
                    }
                    onCheckout={() => checkout.mutate(plan.code)}
                  />
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">
                No plans are available for purchase right now.
              </p>
            )}
          </section>
        )}
      {!canManage &&
        value.deployment_mode === "cloud" &&
        !value.portal_available && (
          <p className="text-sm text-muted-foreground">
            An organization owner or administrator can purchase a Cloud plan.
          </p>
        )}
    </div>
  );
}

function PlanCard({
  plan,
  disabled,
  pending,
  onCheckout,
}: {
  plan: AvailablePlan;
  disabled: boolean;
  pending: boolean;
  onCheckout: () => void;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{plan.name}</CardTitle>
        <CardDescription>{formatPrice(plan)}</CardDescription>
        {plan.description && (
          <CardDescription>{plan.description}</CardDescription>
        )}
      </CardHeader>
      <CardContent className="flex-1">
        <ul className="space-y-2 text-sm text-muted-foreground">
          <li>
            {formatCapacity(plan.limits.max_concurrent_calls ?? null)}{" "}
            concurrent calls
          </li>
          <li>
            {formatCapacity(plan.limits.monthly_runtime_minutes ?? null)}{" "}
            runtime minutes per month
          </li>
          <li>
            {formatCapacity(plan.limits.max_voice_agents ?? null)} voice agents
          </li>
        </ul>
      </CardContent>
      <CardFooter>
        <Button className="w-full" disabled={disabled} onClick={onCheckout}>
          {pending ? "Opening checkout…" : `Choose ${plan.name}`}
        </Button>
      </CardFooter>
    </Card>
  );
}

function RequestError({
  error,
  onRetry,
}: {
  error: unknown;
  onRetry: () => void;
}) {
  const forbidden = error instanceof ApiError && error.status === 403;

  return (
    <div role="alert" className="space-y-3 rounded-lg border p-4">
      <p className="text-sm text-destructive">{billingErrorMessage(error)}</p>
      {!forbidden && (
        <Button variant="outline" onClick={onRetry}>
          Try again
        </Button>
      )}
    </div>
  );
}

function billingErrorMessage(error: unknown): string {
  if (error instanceof ApiError && error.status === 403) {
    return "You do not have permission to manage billing for this organization. Contact an organization owner or administrator.";
  }

  return errorMessage(error);
}

function billingReturnURL(organizationID: string): string {
  const url = new URL("/settings/billing", window.location.origin);
  url.searchParams.set("organization", organizationID);

  return url.href;
}

function redirectToBilling(value: string) {
  const url = new URL(value);
  if (url.protocol !== "https:") {
    throw new ApiError(
      "The billing link is invalid. Please try again.",
      502,
      "INVALID_BILLING_URL",
    );
  }

  window.location.assign(url.href);
}
