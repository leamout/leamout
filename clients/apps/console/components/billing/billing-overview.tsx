import { Badge } from "@leamout/ui/components/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@leamout/ui/components/card";
import {
  formatCapacity,
  formatDate,
  formatPrice,
  runtimeUsagePercent,
} from "./presentation";
import type { BillingStatus } from "./types";

export function BillingOverview({ status }: { status: BillingStatus }) {
  return (
    <div className="space-y-6">
      {status.deployment_mode === "self_hosted" ? (
        <SelfHostedLicense status={status} />
      ) : (
        <CloudSubscription status={status} />
      )}
      {status.deployment_mode === "cloud" ? (
        <RuntimeUsage status={status} />
      ) : (
        <Card>
          <CardHeader>
            <CardTitle>Runtime usage</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              No monthly runtime minute quota.
            </p>
          </CardContent>
        </Card>
      )}
      <CapacityLimits status={status} />
    </div>
  );
}

function CloudSubscription({ status }: { status: BillingStatus }) {
  const subscription = status.subscription;
  const pastDue = subscription?.status === "past_due";
  const graceActive = Boolean(
    subscription?.grace_ends_at &&
      new Date(subscription.grace_ends_at).getTime() > Date.now(),
  );

  return (
    <Card>
      <CardHeader>
        <CardDescription>Current effective plan</CardDescription>
        <CardTitle className="text-xl">
          {status.plan?.name ?? "No plan available"}
        </CardTitle>
        {status.plan && (
          <CardDescription>{formatPrice(status.plan)}</CardDescription>
        )}
      </CardHeader>
      <CardContent className="space-y-4">
        <Badge variant={pastDue ? "destructive" : "secondary"}>
          {subscription?.status.replaceAll("_", " ") ?? "No subscription"}
        </Badge>
        {subscription?.provider && subscription.current_period_end && (
          <p className="text-sm text-muted-foreground">
            {subscription.cancel_at_period_end
              ? "Cancellation scheduled for"
              : subscription.status === "canceled"
                ? "Last billing period ends"
                : subscription.status === "active"
                  ? "Renews on"
                  : subscription.status === "trialing"
                    ? "Trial period ends"
                    : "Current billing period ends"}{" "}
            {formatDate(subscription.current_period_end)}.
          </p>
        )}
        {pastDue && (
          <p role="status" className="text-sm text-destructive">
            {graceActive
              ? `Payment is overdue. Paid limits remain available until ${formatDate(subscription?.grace_ends_at)}. Update your payment details before the grace period ends.`
              : "Payment is overdue. Free limits apply until payment is recovered. Update your payment details in the customer portal."}
          </p>
        )}
        {!status.billing_configured && (
          <p className="text-sm text-muted-foreground">
            Purchases and payment updates are temporarily unavailable. Your
            current limits are shown below.
          </p>
        )}
      </CardContent>
    </Card>
  );
}

function SelfHostedLicense({ status }: { status: BillingStatus }) {
  const license = status.license;
  const claims = license?.license;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Self-hosted license</CardTitle>
        <CardDescription>
          Your commercial license governs this deployment’s capacity and
          features. Cloud subscriptions and monthly runtime quotas do not apply.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <Badge variant={license?.state === "active" ? "secondary" : "outline"}>
          {license?.state.replaceAll("_", " ") ?? "Unavailable"}
        </Badge>
        {claims ? (
          <>
            <dl className="grid gap-4 sm:grid-cols-2">
              <Detail label="License" value={claims.license_id} />
              <Detail label="Deployment" value={claims.deployment_id} />
              <Detail
                label="Licensed nodes"
                value={formatCapacity(claims.max_nodes)}
              />
              <Detail
                label="Licensed concurrent calls"
                value={formatCapacity(claims.max_concurrent_calls)}
              />
              <Detail label="Expires" value={formatDate(claims.expires_at)} />
              <Detail
                label="Grace period ends"
                value={formatDate(claims.grace_until)}
              />
            </dl>
            {claims.features.length > 0 && (
              <div className="space-y-2">
                <p className="text-sm text-muted-foreground">
                  Licensed features
                </p>
                <ul className="flex flex-wrap gap-2">
                  {claims.features.map((feature) => (
                    <li key={feature}>
                      <Badge variant="outline">
                        {feature.replaceAll("_", " ")}
                      </Badge>
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </>
        ) : (
          <p className="text-sm text-muted-foreground">
            No verified license details are available. Contact your deployment
            administrator to install a valid commercial license.
          </p>
        )}
        {license?.state === "grace" && (
          <p role="status" className="text-sm text-muted-foreground">
            Renew your license before the grace period ends to keep starting new
            calls.
          </p>
        )}
        {license?.state === "expired" && (
          <p role="status" className="text-sm text-destructive">
            The license has expired. Install a renewed license to restore new
            calls and licensed features.
          </p>
        )}
      </CardContent>
    </Card>
  );
}

function RuntimeUsage({ status }: { status: BillingStatus }) {
  const usage = status.usage;
  const percent = runtimeUsagePercent(status);
  const usedMinutes = (usage.used_seconds / 60).toLocaleString(undefined, {
    maximumFractionDigits: 1,
  });
  const allowanceMinutes = (usage.allowance_seconds / 60).toLocaleString();

  return (
    <Card>
      <CardHeader>
        <CardTitle>Runtime usage</CardTitle>
        <CardDescription>
          {formatDate(usage.period_start)} – {formatDate(usage.period_end)}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        <p className="text-lg font-medium">
          {usedMinutes} minutes used
          {!usage.unlimited && ` of ${allowanceMinutes}`}
        </p>
        {percent !== null && (
          <progress
            className="h-2 w-full accent-primary"
            max={100}
            value={percent}
            aria-label="Monthly runtime minutes used"
          />
        )}
        <p className="text-sm text-muted-foreground">
          {usage.unlimited
            ? "No monthly runtime minute quota."
            : "Usage includes runtime across your organization’s calls."}
        </p>
      </CardContent>
    </Card>
  );
}

function CapacityLimits({ status }: { status: BillingStatus }) {
  const limits = status.limits;
  const retention = limits.retention_days;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Effective limits</CardTitle>
      </CardHeader>
      <CardContent>
        <dl className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
          <Detail
            label="Concurrent calls"
            value={formatCapacity(limits.max_concurrent_calls)}
          />
          <Detail
            label="Voice agents"
            value={formatCapacity(limits.max_voice_agents)}
          />
          <Detail
            label="SIP trunks"
            value={formatCapacity(limits.max_sip_trunks)}
          />
          <Detail
            label="Monthly runtime minutes"
            value={formatCapacity(limits.monthly_runtime_minutes)}
          />
          <Detail
            label="Recording retention"
            value={
              retention === null
                ? "No plan restriction"
                : `${retention.toLocaleString()} days`
            }
          />
        </dl>
      </CardContent>
    </Card>
  );
}

function Detail({ label, value }: { label: string; value: string }) {
  return (
    <div className="space-y-1">
      <dt className="text-sm text-muted-foreground">{label}</dt>
      <dd className="break-words font-medium">{value}</dd>
    </div>
  );
}
