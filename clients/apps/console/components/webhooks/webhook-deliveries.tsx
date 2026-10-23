"use client";

import { Button } from "@leamout/ui/components/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@leamout/ui/components/table";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { api, errorMessage } from "@/lib/api";
import type { WebhookDelivery } from "./types";

const pageSize = 50;

export function WebhookDeliveries({ webhookId }: { webhookId: string }) {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const client = useQueryClient();
  const [offset, setOffset] = useState(0);
  const [retrying, setRetrying] = useState<string | null>(null);
  const base = "/webhooks/" + encodeURIComponent(webhookId) + "/deliveries";
  const key = ["webhook-deliveries", organizationId, webhookId, offset];
  const deliveries = useQuery({
    queryKey: key,
    queryFn: ({ signal }) => api<{ deliveries: WebhookDelivery[] }>(
      base + "?limit=" + pageSize + "&offset=" + offset,
      { organizationId, signal },
    ),
    enabled: Boolean(organizationId),
  });
  const retry = useMutation({
    mutationFn: (id: string) => api<WebhookDelivery>(
      base + "/" + encodeURIComponent(id) + "/retry",
      { method: "POST", organizationId },
    ),
    onSuccess: async () => {
      setRetrying(null);
      await client.invalidateQueries({
        queryKey: ["webhook-deliveries", organizationId, webhookId],
      });
    },
  });
  const items = deliveries.data?.deliveries ?? [];

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-lg font-semibold">Delivery history</h2>
        <Button variant="outline" disabled={deliveries.isFetching}
          onClick={() => void deliveries.refetch()}>Refresh</Button>
      </div>
      {deliveries.isPending ? <p role="status">Loading deliveries…</p> : deliveries.isError ? (
        <div role="alert" className="space-y-3">
          <p>{errorMessage(deliveries.error)}</p>
          <Button variant="outline" onClick={() => void deliveries.refetch()}>Retry</Button>
        </div>
      ) : (
        <>
          <div className="overflow-x-auto rounded-lg border">
            <Table>
              <TableHeader><TableRow>
                <TableHead>Delivery / Event</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Attempts</TableHead>
                <TableHead>Replays</TableHead>
                <TableHead>Response</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow></TableHeader>
              <TableBody>
                {items.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell>
                      <details className="max-w-md">
                        <summary className="cursor-pointer break-all font-mono text-xs">{item.id}</summary>
                        <div className="mt-3 space-y-2 text-xs">
                          <p>Event: {item.event_id}</p>
                          <p>Last attempt: {item.last_attempt_at ?? "—"}</p>
                          <p>Error: {item.last_error ?? "None"}</p>
                          <pre className="max-h-48 overflow-auto whitespace-pre-wrap break-all">{item.response_body ?? "No response body"}</pre>
                        </div>
                      </details>
                    </TableCell>
                    <TableCell>{item.status}</TableCell>
                    <TableCell>{item.attempt_count}</TableCell>
                    <TableCell>{item.replay_count}</TableCell>
                    <TableCell>{item.response_status ?? "—"}</TableCell>
                    <TableCell>
                      {retrying === item.id ? (
                        <div className="flex gap-2">
                          <Button variant="outline" disabled={retry.isPending}
                            onClick={() => retry.mutate(item.id)}>Confirm retry</Button>
                          <Button variant="outline" onClick={() => setRetrying(null)}>Cancel</Button>
                        </div>
                      ) : (
                        <Button variant="outline" onClick={() => setRetrying(item.id)}>Retry…</Button>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
                {!items.length && <TableRow><TableCell colSpan={6} className="py-12 text-center">No deliveries on this page.</TableCell></TableRow>}
              </TableBody>
            </Table>
          </div>
          <div className="flex items-center justify-between gap-3">
            <p className="text-sm text-muted-foreground">Page {offset / pageSize + 1}</p>
            <div className="flex gap-2">
              <Button variant="outline" disabled={offset === 0} onClick={() => setOffset((n) => Math.max(0, n - pageSize))}>Previous</Button>
              <Button variant="outline" disabled={items.length < pageSize} onClick={() => setOffset((n) => n + pageSize)}>Next</Button>
            </div>
          </div>
        </>
      )}
      {retry.isError && <p role="alert" className="text-destructive">{errorMessage(retry.error)}</p>}
    </div>
  );
}
