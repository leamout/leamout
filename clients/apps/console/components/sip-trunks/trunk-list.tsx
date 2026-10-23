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
import type { SipTrunk } from "./types";

export function TrunkList() {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState("all");

  const trunks = useQuery({
    queryKey: ["trunks", organizationId],
    queryFn: ({ signal }) =>
      api<{ trunks: SipTrunk[] }>("/trunks/", {
        organizationId,
        signal,
      }),
    enabled: Boolean(organizationId),
  });

  const items = trunks.data?.trunks ?? [];
  const filtered = items.filter(
    (trunk) =>
      trunk.name.toLowerCase().includes(search.trim().toLowerCase()) &&
      (status === "all" || trunk.status === status),
  );

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap justify-between gap-4">
        <div>
          <h1 className="text-3xl font-semibold tracking-tight">SIP trunks</h1>
          <p className="mt-2 text-muted-foreground">
            Connect and manage your own SIP carrier or PBX.
          </p>
        </div>
        <Button render={<Link href="/sip-trunks/new" />}>
          Create SIP trunk
        </Button>
      </div>

      <div className="flex flex-wrap gap-4">
        <div className="flex-1 space-y-2">
          <label htmlFor="trunk-search" className="text-sm font-medium">
            Search trunks
          </label>
          <Input
            id="trunk-search"
            type="search"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
          />
        </div>
        <div className="space-y-2">
          <label htmlFor="trunk-status" className="block text-sm font-medium">
            Status
          </label>
          <select
            id="trunk-status"
            className="h-9 rounded-md border bg-background px-3 text-sm"
            value={status}
            onChange={(event) => setStatus(event.target.value)}
          >
            <option value="all">All</option>
            <option value="active">Active</option>
            <option value="disabled">Disabled</option>
          </select>
        </div>
      </div>

      {trunks.isPending ? (
        <p role="status">Loading SIP trunks…</p>
      ) : trunks.isError ? (
        <div role="alert" className="space-y-3">
          <p>{errorMessage(trunks.error)}</p>
          <Button variant="outline" onClick={() => void trunks.refetch()}>
            Retry
          </Button>
        </div>
      ) : (
        <div className="overflow-hidden rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                {["Name", "Direction", "Status", "Inbound", "Updated"].map(
                  (label) => (
                    <TableHead key={label}>{label}</TableHead>
                  ),
                )}
              </TableRow>
            </TableHeader>
            <TableBody>
              {filtered.map((trunk) => (
                <TableRow key={trunk.id}>
                  <TableCell>
                    <Link
                      className="font-medium hover:underline"
                      href={`/sip-trunks/${encodeURIComponent(trunk.id)}`}
                    >
                      {trunk.name}
                    </Link>
                  </TableCell>
                  <TableCell>{trunk.direction}</TableCell>
                  <TableCell>{trunk.status}</TableCell>
                  <TableCell>
                    {trunk.inbound_enabled ? "Enabled" : "Disabled"}
                  </TableCell>
                  <TableCell>
                    {new Date(trunk.updated_at).toLocaleDateString()}
                  </TableCell>
                </TableRow>
              ))}
              {filtered.length === 0 && (
                <TableRow>
                  <TableCell
                    colSpan={5}
                    className="py-12 text-center text-muted-foreground"
                  >
                    {items.length
                      ? "No matching trunks."
                      : "No SIP trunks yet. Create one to connect your carrier."}
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      )}
      <p className="text-xs text-muted-foreground">
        An active trunk is not necessarily healthy. Use Validate on trunk
        details to check configuration.
      </p>
    </div>
  );
}
