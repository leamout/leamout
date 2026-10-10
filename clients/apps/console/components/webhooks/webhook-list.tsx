"use client";

import { Button } from "@leamout/ui/components/button";
import { Input } from "@leamout/ui/components/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@leamout/ui/components/table";
import Link from "next/link";
import { useState } from "react";
import type { WebhookEndpoint } from "@/components/webhooks/types";

export function WebhookList({
  endpoints = [],
}: {
  endpoints?: WebhookEndpoint[];
}) {
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState("all");
  const filtered = endpoints.filter(
    (endpoint) =>
      `${endpoint.url} ${endpoint.subscribedEvents.join(" ")}`
        .toLowerCase()
        .includes(search.toLowerCase()) &&
      (status === "all" || endpoint.enabled === (status === "enabled")),
  );
  return (
    <div className="space-y-6">
      <div className="flex flex-wrap justify-between gap-4">
        <div className="space-y-2">
          <h1 className="text-3xl font-semibold tracking-tight">Webhooks</h1>
          <p className="text-muted-foreground">
            Deliver platform events to your application.
          </p>
        </div>
        <Button render={<Link href="/webhooks/new" />}>Create endpoint</Button>
      </div>
      <div className="flex flex-wrap gap-4">
        <div className="flex-1 space-y-2">
          <label htmlFor="webhook-search" className="text-sm font-medium">
            Search endpoints
          </label>
          <Input
            id="webhook-search"
            type="search"
            placeholder="URL or event name"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
          />
        </div>
        <div className="space-y-2">
          <label htmlFor="webhook-status" className="block text-sm font-medium">
            Status
          </label>
          <select
            id="webhook-status"
            className="h-9 rounded-md border bg-background px-3"
            value={status}
            onChange={(event) => setStatus(event.target.value)}
          >
            <option value="all">All statuses</option>
            <option value="enabled">Enabled</option>
            <option value="disabled">Disabled</option>
          </select>
        </div>
      </div>
      <div className="overflow-hidden rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              {["Endpoint", "Status", "Events", "Consecutive failures"].map(
                (label) => (
                  <TableHead key={label}>{label}</TableHead>
                ),
              )}
            </TableRow>
          </TableHeader>
          <TableBody>
            {filtered.map((endpoint) => (
              <TableRow key={endpoint.id}>
                <TableCell>
                  <Link
                    className="font-medium hover:underline"
                    href={`/webhooks/${encodeURIComponent(endpoint.id)}`}
                  >
                    {endpoint.url}
                  </Link>
                </TableCell>
                <TableCell>
                  {endpoint.enabled ? "Enabled" : "Disabled"}
                </TableCell>
                <TableCell>{endpoint.subscribedEvents.join(", ")}</TableCell>
                <TableCell>{endpoint.consecutiveFailures}</TableCell>
              </TableRow>
            ))}
            {filtered.length === 0 && (
              <TableRow>
                <TableCell colSpan={4} className="h-40 text-center">
                  <p className="font-medium">
                    {endpoints.length
                      ? "No matching endpoints"
                      : "No webhook endpoints to display"}
                  </p>
                  <p className="mt-2 text-muted-foreground">
                    {endpoints.length
                      ? "Try another search or filter."
                      : "Endpoints will appear here when account integration is connected."}
                  </p>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <p className="text-sm text-muted-foreground">
        UI preview. Event delivery is not connected.
      </p>
    </div>
  );
}
