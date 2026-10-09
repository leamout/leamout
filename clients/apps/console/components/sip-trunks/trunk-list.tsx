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

type Trunk = {
  id: string;
  name: string;
  direction: string;
  status: "active" | "disabled";
  endpoints: number;
  updated: string;
};

export function TrunkList({ trunks = [] }: { trunks?: Trunk[] }) {
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState("all");
  const filtered = trunks.filter(
    (trunk) =>
      trunk.name.toLowerCase().includes(search.toLowerCase()) &&
      (status === "all" || trunk.status === status),
  );
  return (
    <div className="space-y-6">
      <div className="flex flex-wrap justify-between gap-4">
        <div className="space-y-2">
          <h1 className="text-3xl font-semibold tracking-tight">SIP trunks</h1>
          <p className="text-muted-foreground">
            Connect your carrier or PBX to Leamout.
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
            className="h-9 rounded-md border bg-background px-3"
            value={status}
            onChange={(event) => setStatus(event.target.value)}
          >
            <option value="all">All statuses</option>
            <option value="active">Active</option>
            <option value="disabled">Disabled</option>
          </select>
        </div>
      </div>
      <div className="overflow-hidden rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              {["Name", "Direction", "Status", "Endpoints", "Updated"].map(
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
                <TableCell>{trunk.endpoints}</TableCell>
                <TableCell>{trunk.updated}</TableCell>
              </TableRow>
            ))}
            {filtered.length === 0 && (
              <TableRow>
                <TableCell colSpan={5} className="h-40 text-center">
                  <p className="font-medium">
                    {trunks.length
                      ? "No matching trunks"
                      : "No SIP trunks to display"}
                  </p>
                  <p className="mt-2 text-muted-foreground">
                    {trunks.length
                      ? "Try another search or filter."
                      : "Trunks will appear here when account integration is connected."}
                  </p>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <p className="text-sm text-muted-foreground">
        UI preview. Active status does not imply endpoint health.
      </p>
    </div>
  );
}
