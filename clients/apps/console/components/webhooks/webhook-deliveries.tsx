"use client";

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@leamout/ui/components/table";
import { useState } from "react";
import type { WebhookDelivery } from "@/components/webhooks/types";

export function WebhookDeliveries({
  deliveries = [],
}: {
  deliveries?: WebhookDelivery[];
}) {
  const [status, setStatus] = useState("all");
  const statuses = [...new Set(deliveries.map((delivery) => delivery.status))];
  const filtered = deliveries.filter(
    (delivery) => status === "all" || delivery.status === status,
  );
  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <label htmlFor="delivery-status" className="block text-sm font-medium">
          Delivery status
        </label>
        <select
          id="delivery-status"
          className="h-9 rounded-md border bg-background px-3"
          value={status}
          onChange={(event) => setStatus(event.target.value)}
        >
          <option value="all">All statuses</option>
          {statuses.map((value) => (
            <option key={value} value={value}>
              {value}
            </option>
          ))}
        </select>
      </div>
      <div className="overflow-hidden rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              {[
                "Delivery / Event",
                "Status",
                "Attempts",
                "Replays",
                "Last attempt",
                "HTTP response",
              ].map((label) => (
                <TableHead key={label}>{label}</TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {filtered.map((delivery) => (
              <TableRow key={delivery.id}>
                <TableCell>
                  <details>
                    <summary className="cursor-pointer font-medium">
                      {delivery.id}
                    </summary>
                    <div className="mt-3 max-w-md space-y-2 whitespace-normal">
                      <p>Event: {delivery.eventId}</p>
                      <p>Error: {delivery.lastError ?? "None reported"}</p>
                      <pre className="max-h-64 overflow-auto whitespace-pre-wrap break-all text-xs">
                        {delivery.responseBody ?? "No response body available."}
                      </pre>
                    </div>
                  </details>
                </TableCell>
                <TableCell>{delivery.status}</TableCell>
                <TableCell>{delivery.attemptCount}</TableCell>
                <TableCell>{delivery.replayCount}</TableCell>
                <TableCell>{delivery.lastAttemptAt ?? "—"}</TableCell>
                <TableCell>{delivery.responseStatus ?? "—"}</TableCell>
              </TableRow>
            ))}
            {filtered.length === 0 && (
              <TableRow>
                <TableCell colSpan={6} className="h-32 text-center">
                  {deliveries.length
                    ? "No deliveries match this status."
                    : "Delivery history has not been loaded."}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <p className="text-sm text-muted-foreground">
        Expand a delivery to inspect its event ID, error, and response body.
        Replay actions will be connected later.
      </p>
    </div>
  );
}
