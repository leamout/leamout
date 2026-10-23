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
import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { useState } from "react";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { api, errorMessage } from "@/lib/api";
import type { WebhookEndpoint } from "./types";

export function WebhookList() {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState("all");
  const endpoints = useQuery({
    queryKey: ["webhooks", organizationId],
    queryFn: ({ signal }) =>
      api<{ webhooks: WebhookEndpoint[] }>("/webhooks/", {
        organizationId,
        signal,
      }),
    enabled: Boolean(organizationId),
  });
  const items = endpoints.data?.webhooks ?? [];
  const filtered = items.filter(
    (item) =>
      `${item.url} ${item.subscribed_events.join(" ")}`
        .toLowerCase()
        .includes(search.trim().toLowerCase()) &&
      (status === "all" || item.enabled === (status === "enabled")),
  );

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap justify-between gap-4">
        <div className="space-y-2">
          <h1 className="text-3xl font-semibold tracking-tight">Webhooks</h1>
          <p className="text-muted-foreground">
            Send organization platform events to your HTTPS endpoints.
          </p>
        </div>
        <Button render={<Link href="/webhooks/new" />}>Create endpoint</Button>
      </div>
      <div className="flex flex-wrap gap-4">
        <label className="flex-1 space-y-2 text-sm">
          <span className="block font-medium">Search endpoints</span>
          <Input
            type="search"
            aria-label="Search endpoints"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
          />
        </label>
        <label className="space-y-2 text-sm">
          <span className="block font-medium">Status</span>
          <select
            className="block h-9 rounded-md border bg-background px-3"
            value={status}
            onChange={(event) => setStatus(event.target.value)}
          >
            <option value="all">All</option>
            <option value="enabled">Enabled</option>
            <option value="disabled">Disabled</option>
          </select>
        </label>
      </div>
      {endpoints.isPending ? (
        <p role="status">Loading webhooks…</p>
      ) : endpoints.isError ? (
        <div role="alert" className="space-y-3">
          <p>{errorMessage(endpoints.error)}</p>
          <Button variant="outline" onClick={() => void endpoints.refetch()}>
            Retry
          </Button>
        </div>
      ) : (
        <div className="overflow-x-auto rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Endpoint</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Subscribed events</TableHead>
                <TableHead>Failures</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filtered.map((endpoint) => (
                <TableRow key={endpoint.id}>
                  <TableCell>
                    <Link
                      className="break-all font-medium hover:underline"
                      href={`/webhooks/${encodeURIComponent(endpoint.id)}`}
                    >
                      {endpoint.url}
                    </Link>
                  </TableCell>
                  <TableCell>
                    {endpoint.enabled ? "Enabled" : "Disabled"}
                  </TableCell>
                  <TableCell>{endpoint.subscribed_events.join(", ")}</TableCell>
                  <TableCell>{endpoint.consecutive_failures}</TableCell>
                </TableRow>
              ))}
              {filtered.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4} className="py-12 text-center">
                    {items.length
                      ? "No matching endpoints."
                      : "No webhooks configured."}
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  );
}
