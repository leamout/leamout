"use client";

import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@leamout/ui/components/tabs";
import Link from "next/link";
import type {
  WebhookDelivery,
  WebhookEndpoint,
} from "@/components/webhooks/types";
import { WebhookDeliveries } from "@/components/webhooks/webhook-deliveries";
import { WebhookForm } from "@/components/webhooks/webhook-form";

export function WebhookDetails({
  webhookId,
  endpoint,
  deliveries,
}: {
  webhookId: string;
  endpoint?: WebhookEndpoint;
  deliveries?: WebhookDelivery[];
}) {
  return (
    <div className="space-y-6">
      <Link
        href="/webhooks"
        className="text-sm text-muted-foreground hover:underline"
      >
        Back to webhooks
      </Link>
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">
          Webhook endpoint
        </h1>
        <p className="break-all text-sm text-muted-foreground">
          Endpoint ID: {webhookId}
        </p>
        <p className="text-sm text-muted-foreground">
          UI preview.{" "}
          {endpoint
            ? "Preview configuration is shown below."
            : "Endpoint configuration has not been loaded."}
        </p>
      </div>
      <dl className="grid gap-6 rounded-lg border p-6 sm:grid-cols-2">
        <div>
          <dt className="text-sm text-muted-foreground">Status</dt>
          <dd className="mt-1 font-medium">
            {endpoint
              ? endpoint.enabled
                ? "Enabled"
                : "Disabled"
              : "Unavailable"}
          </dd>
        </div>
        <div>
          <dt className="text-sm text-muted-foreground">
            Consecutive failures
          </dt>
          <dd className="mt-1 font-medium">
            {endpoint?.consecutiveFailures ?? "—"}
          </dd>
        </div>
      </dl>
      {endpoint?.disabledReason && (
        <p className="rounded-lg border p-4 text-sm">
          Disabled reason: {endpoint.disabledReason}
        </p>
      )}
      <Tabs defaultValue="configuration" className="gap-6">
        <TabsList aria-label="Webhook endpoint details">
          <TabsTrigger value="configuration">Configuration</TabsTrigger>
          <TabsTrigger value="deliveries">Deliveries</TabsTrigger>
          <TabsTrigger value="signing">Signing</TabsTrigger>
        </TabsList>
        <TabsContent value="configuration">
          <WebhookForm endpoint={endpoint} />
        </TabsContent>
        <TabsContent value="deliveries">
          <WebhookDeliveries deliveries={deliveries} />
        </TabsContent>
        <TabsContent value="signing">
          <div className="space-y-3 rounded-lg border p-6">
            <h2 className="font-semibold">Signing secret</h2>
            <p className="text-sm text-muted-foreground">
              Secret creation and rotation will be connected later. No signing
              secret is generated or displayed in this preview.
            </p>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
