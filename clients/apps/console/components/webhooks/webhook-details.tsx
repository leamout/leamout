"use client";

import { Button } from "@leamout/ui/components/button";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@leamout/ui/components/tabs";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { api, errorMessage } from "@/lib/api";
import { SecretReveal } from "./secret-reveal";
import type { WebhookEndpoint, WebhookSecretResponse } from "./types";
import { WebhookDeliveries } from "./webhook-deliveries";
import { WebhookForm } from "./webhook-form";

export function WebhookDetails({ webhookId }: { webhookId: string }) {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const client = useQueryClient();
  const router = useRouter();
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [confirmRotate, setConfirmRotate] = useState(false);
  const path = "/webhooks/" + encodeURIComponent(webhookId);
  const key = ["webhooks", organizationId, webhookId];

  const endpoint = useQuery({
    queryKey: key,
    queryFn: ({ signal }) =>
      api<WebhookEndpoint>(path, {
        organizationId,
        signal,
      }),
    enabled: Boolean(organizationId),
  });

  const test = useMutation({
    mutationFn: () =>
      api<{ response_status: number }>(path + "/test", {
        method: "POST",
        organizationId,
      }),
  });

  const rotate = useMutation({
    mutationFn: () =>
      api<WebhookSecretResponse>(path + "/rotate-secret", {
        method: "POST",
        organizationId,
      }),
    onSuccess: async (result) => {
      setConfirmRotate(false);
      client.setQueryData(key, result.webhook);
      await client.invalidateQueries({
        queryKey: ["webhooks", organizationId],
      });
    },
  });

  const remove = useMutation({
    mutationFn: () =>
      api<void>(path, {
        method: "DELETE",
        organizationId,
      }),
    onSuccess: async () => {
      client.removeQueries({ queryKey: key, exact: true });
      await client.invalidateQueries({
        queryKey: ["webhooks", organizationId],
      });
      router.replace("/webhooks");
    },
  });

  if (endpoint.isPending) {
    return <p role="status">Loading webhook…</p>;
  }

  if (endpoint.isError) {
    return (
      <div role="alert" className="space-y-3">
        <p>{errorMessage(endpoint.error)}</p>
        <Button variant="outline" onClick={() => void endpoint.refetch()}>
          Retry
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-7">
      <Link
        href="/webhooks"
        className="text-sm text-muted-foreground hover:underline"
      >
        Back to webhooks
      </Link>

      <div className="space-y-2">
        <h1 className="text-3xl font-semibold">Webhook endpoint</h1>
        <p className="break-all text-muted-foreground">
          {endpoint.data.url}
        </p>
        <p className="font-mono text-xs text-muted-foreground">
          {endpoint.data.id}
        </p>
        <p className="text-sm text-muted-foreground">
          {endpoint.data.enabled ? "Enabled" : "Disabled"} ·{" "}
          {endpoint.data.consecutive_failures} consecutive failures
        </p>
        {endpoint.data.disabled_reason && (
          <p role="alert">{endpoint.data.disabled_reason}</p>
        )}
      </div>

      <Tabs defaultValue="configuration" className="gap-5">
        <TabsList>
          <TabsTrigger value="configuration">Configuration</TabsTrigger>
          <TabsTrigger value="deliveries">Deliveries</TabsTrigger>
          <TabsTrigger value="signing">Signing & testing</TabsTrigger>
        </TabsList>

        <TabsContent value="configuration">
          <WebhookForm endpoint={endpoint.data} key={endpoint.data.id} />
        </TabsContent>

        <TabsContent value="deliveries">
          <WebhookDeliveries webhookId={webhookId} />
        </TabsContent>

        <TabsContent value="signing">
          <section className="space-y-4">
            <h2 className="text-lg font-semibold">Test delivery</h2>
            <p className="text-sm text-muted-foreground">
              Sends a signed test payload to this endpoint.
            </p>
            <Button
              variant="outline"
              disabled={test.isPending || !endpoint.data.enabled}
              onClick={() => test.mutate()}
            >
              Send test
            </Button>
            {test.isSuccess && (
              <p role="status">
                Endpoint returned HTTP {test.data.response_status}.
              </p>
            )}
            {test.isError && (
              <p role="alert">{errorMessage(test.error)}</p>
            )}

            <h2 className="text-lg font-semibold">Signing secret</h2>
            <p className="text-sm text-muted-foreground">
              Rotation invalidates the previous secret. Copy the replacement
              before leaving this page.
            </p>
            {rotate.data && (
              <SecretReveal
                secret={rotate.data.signing_secret}
                label="New signing secret"
              />
            )}

            {!confirmRotate ? (
              <Button
                variant="outline"
                onClick={() => setConfirmRotate(true)}
              >
                Rotate signing secret…
              </Button>
            ) : (
              <div className="flex gap-2">
                <Button
                  variant="destructive"
                  disabled={rotate.isPending}
                  onClick={() => rotate.mutate()}
                >
                  Confirm rotation
                </Button>
                <Button
                  variant="outline"
                  onClick={() => setConfirmRotate(false)}
                >
                  Cancel
                </Button>
              </div>
            )}
            {rotate.isError && (
              <p role="alert">{errorMessage(rotate.error)}</p>
            )}
          </section>
        </TabsContent>
      </Tabs>

      <section className="space-y-3 border-t pt-6">
        <h2 className="font-semibold">Disable endpoint</h2>
        <p className="text-sm text-muted-foreground">
          Removing an endpoint disables it and cancels pending deliveries.
        </p>
        {!confirmDelete ? (
          <Button
            variant="outline"
            onClick={() => setConfirmDelete(true)}
          >
            Remove endpoint…
          </Button>
        ) : (
          <div className="flex gap-2">
            <Button
              variant="destructive"
              disabled={remove.isPending}
              onClick={() => remove.mutate()}
            >
              Confirm removal
            </Button>
            <Button
              variant="outline"
              onClick={() => setConfirmDelete(false)}
            >
              Cancel
            </Button>
          </div>
        )}
        {remove.isError && (
          <p role="alert">{errorMessage(remove.error)}</p>
        )}
      </section>
    </div>
  );
}
