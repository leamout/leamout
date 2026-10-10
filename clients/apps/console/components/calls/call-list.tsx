"use client";

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
import { type CallSummary, formatDuration } from "@/components/calls/types";

export function CallList({ calls = [] }: { calls?: CallSummary[] }) {
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState("all");
  const [direction, setDirection] = useState("all");
  const filtered = calls.filter(
    (call) =>
      `${call.id} ${call.agentName} ${call.from} ${call.to}`
        .toLowerCase()
        .includes(search.toLowerCase()) &&
      (status === "all" || call.status === status) &&
      (direction === "all" || call.direction === direction),
  );

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">Calls</h1>
        <p className="text-muted-foreground">
          Review conversations, call outcomes, and runtime diagnostics.
        </p>
      </div>
      <div className="flex flex-wrap gap-4">
        <div className="min-w-48 flex-1 space-y-2">
          <label htmlFor="call-search" className="text-sm font-medium">
            Search calls
          </label>
          <Input
            id="call-search"
            type="search"
            placeholder="Call ID, agent, or phone number"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
          />
        </div>
        <div className="space-y-2">
          <label htmlFor="call-status" className="block text-sm font-medium">
            Status
          </label>
          <select
            id="call-status"
            className="h-9 rounded-md border bg-background px-3 text-sm"
            value={status}
            onChange={(event) => setStatus(event.target.value)}
          >
            <option value="all">All statuses</option>
            {[
              "queued",
              "ringing",
              "active",
              "completed",
              "failed",
              "canceled",
            ].map((value) => (
              <option key={value} value={value}>
                {value}
              </option>
            ))}
          </select>
        </div>
        <div className="space-y-2">
          <label htmlFor="call-direction" className="block text-sm font-medium">
            Direction
          </label>
          <select
            id="call-direction"
            className="h-9 rounded-md border bg-background px-3 text-sm"
            value={direction}
            onChange={(event) => setDirection(event.target.value)}
          >
            <option value="all">All directions</option>
            <option value="inbound">Inbound</option>
            <option value="outbound">Outbound</option>
          </select>
        </div>
      </div>
      <div className="overflow-hidden rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              {[
                "Call",
                "Agent",
                "Direction",
                "Status",
                "From / To",
                "Started",
                "Duration",
              ].map((label) => (
                <TableHead key={label}>{label}</TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {filtered.map((call) => (
              <TableRow key={call.id}>
                <TableCell>
                  <Link
                    className="font-medium hover:underline"
                    href={`/calls/${encodeURIComponent(call.id)}`}
                  >
                    {call.id}
                  </Link>
                </TableCell>
                <TableCell>{call.agentName}</TableCell>
                <TableCell className="capitalize">{call.direction}</TableCell>
                <TableCell className="capitalize">{call.status}</TableCell>
                <TableCell>
                  <p>{call.from}</p>
                  <p className="text-muted-foreground">{call.to}</p>
                </TableCell>
                <TableCell>{call.startedAt}</TableCell>
                <TableCell>{formatDuration(call.durationSeconds)}</TableCell>
              </TableRow>
            ))}
            {filtered.length === 0 && (
              <TableRow>
                <TableCell colSpan={7} className="h-40 text-center">
                  <p className="font-medium">
                    {calls.length === 0
                      ? "No calls to display"
                      : "No matching calls"}
                  </p>
                  <p className="mt-2 text-muted-foreground">
                    {calls.length === 0
                      ? "Call history will appear here when runtime integration is connected."
                      : "Try another search or filter."}
                  </p>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <p className="text-sm text-muted-foreground">
        UI preview. Live call data is not connected.
      </p>
    </div>
  );
}
