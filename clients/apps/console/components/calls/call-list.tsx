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
import type { VoiceAgent } from "@/components/agents/types";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { api, errorMessage } from "@/lib/api";
import {
  type Call,
  callDurationSeconds,
  callStates,
  formatCallDate,
  formatDuration,
} from "./types";

const pageSize = 50;

export function CallList() {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const [search, setSearch] = useState("");
  const [state, setState] = useState("all");
  const [direction, setDirection] = useState("all");
  const [agentId, setAgentId] = useState("all");
  const [offset, setOffset] = useState(0);

  const calls = useQuery({
    queryKey: ["calls", organizationId, state, direction, agentId, offset],
    queryFn: ({ signal }) => {
      const params = new URLSearchParams({
        limit: String(pageSize),
        offset: String(offset),
      });
      if (state !== "all") params.set("state", state);
      if (direction !== "all") params.set("direction", direction);
      if (agentId !== "all") params.set("voice_agent_id", agentId);

      return api<{ calls: Call[] }>(`/calls/?${params.toString()}`, {
        organizationId,
        signal,
      });
    },
    enabled: Boolean(organizationId),
  });

  const agents = useQuery({
    queryKey: ["voice-agents", organizationId],
    queryFn: ({ signal }) =>
      api<{ voice_agents: VoiceAgent[] }>("/voice-agents/", {
        organizationId,
        signal,
      }),
    enabled: Boolean(organizationId),
  });

  const items = calls.data?.calls ?? [];
  const agentNames = new Map(
    agents.data?.voice_agents.map((agent) => [agent.id, agent.name] as const) ??
      [],
  );
  const filtered = items.filter((call) =>
    [
      call.id,
      call.from_uri,
      call.to_uri,
      call.sip_call_id ?? "",
      call.voice_agent_id ? (agentNames.get(call.voice_agent_id) ?? "") : "",
    ]
      .join(" ")
      .toLowerCase()
      .includes(search.trim().toLowerCase()),
  );

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="space-y-2">
          <h1 className="text-3xl font-semibold tracking-tight">Calls</h1>
          <p className="text-muted-foreground">
            Review call outcomes and persisted SIP and routing diagnostics.
          </p>
        </div>
        <Button
          variant="outline"
          disabled={calls.isFetching || !organizationId}
          onClick={() => void calls.refetch()}
        >
          {calls.isFetching ? "Refreshing…" : "Refresh"}
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-4">
        <div className="space-y-2">
          <label htmlFor="call-search" className="block text-sm font-medium">
            Search current page
          </label>
          <Input
            id="call-search"
            type="search"
            placeholder="Call ID, SIP ID, or URI"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
          />
        </div>
        <div className="space-y-2">
          <label htmlFor="call-state" className="block text-sm font-medium">
            State
          </label>
          <select
            id="call-state"
            className="h-9 w-full rounded-md border bg-background px-3 text-sm"
            value={state}
            onChange={(event) => {
              setState(event.target.value);
              setOffset(0);
            }}
          >
            <option value="all">All states</option>
            {callStates.map((item) => (
              <option key={item} value={item}>
                {item}
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
            className="h-9 w-full rounded-md border bg-background px-3 text-sm"
            value={direction}
            onChange={(event) => {
              setDirection(event.target.value);
              setOffset(0);
            }}
          >
            <option value="all">All directions</option>
            <option value="inbound">Inbound</option>
            <option value="outbound">Outbound</option>
          </select>
        </div>
        <div className="space-y-2">
          <label htmlFor="call-agent" className="block text-sm font-medium">
            Agent
          </label>
          <select
            id="call-agent"
            className="h-9 w-full rounded-md border bg-background px-3 text-sm"
            value={agentId}
            disabled={agents.isPending || agents.isError}
            onChange={(event) => {
              setAgentId(event.target.value);
              setOffset(0);
            }}
          >
            <option value="all">All Agents</option>
            {agents.data?.voice_agents.map((agent) => (
              <option key={agent.id} value={agent.id}>
                {agent.name}
              </option>
            ))}
          </select>
        </div>
      </div>

      {agents.isError && (
        <p role="alert" className="text-sm text-muted-foreground">
          Agent names are unavailable: {errorMessage(agents.error)}
        </p>
      )}

      {calls.isPending ? (
        <p role="status" className="rounded-lg border p-6">
          Loading calls…
        </p>
      ) : calls.isError ? (
        <div role="alert" className="space-y-3 rounded-lg border p-6">
          <p>{errorMessage(calls.error)}</p>
          <Button variant="outline" onClick={() => void calls.refetch()}>
            Try again
          </Button>
        </div>
      ) : (
        <>
          <div className="overflow-x-auto rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Call</TableHead>
                  <TableHead>Agent</TableHead>
                  <TableHead>Direction</TableHead>
                  <TableHead>State</TableHead>
                  <TableHead>From / To</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead>Talk time</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filtered.map((call) => (
                  <TableRow key={call.id}>
                    <TableCell>
                      <Link
                        className="font-mono text-xs hover:underline"
                        href={`/calls/${encodeURIComponent(call.id)}`}
                      >
                        {call.id}
                      </Link>
                    </TableCell>
                    <TableCell>
                      {call.voice_agent_id ? (
                        <Link
                          className="hover:underline"
                          href={`/agents/${encodeURIComponent(call.voice_agent_id)}`}
                        >
                          {agentNames.get(call.voice_agent_id) ??
                            call.voice_agent_id}
                        </Link>
                      ) : (
                        "—"
                      )}
                    </TableCell>
                    <TableCell className="capitalize">
                      {call.direction}
                    </TableCell>
                    <TableCell className="capitalize">{call.state}</TableCell>
                    <TableCell>
                      <p className="break-all">{call.from_uri}</p>
                      <p className="break-all text-muted-foreground">
                        {call.to_uri}
                      </p>
                    </TableCell>
                    <TableCell>{formatCallDate(call.created_at)}</TableCell>
                    <TableCell>
                      {formatDuration(callDurationSeconds(call))}
                    </TableCell>
                  </TableRow>
                ))}
                {!filtered.length && (
                  <TableRow>
                    <TableCell colSpan={7} className="h-36 text-center">
                      {items.length
                        ? "No calls match this page search."
                        : offset
                          ? "No more calls on this page."
                          : "No calls match these filters."}
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
          <div className="flex flex-wrap items-center justify-between gap-4">
            <p className="text-sm text-muted-foreground">
              Page {Math.floor(offset / pageSize) + 1} · {items.length} calls
              fetched. Search applies to this page only.
            </p>
            <div className="flex gap-2">
              <Button
                variant="outline"
                disabled={offset === 0 || calls.isFetching}
                onClick={() =>
                  setOffset((value) => Math.max(0, value - pageSize))
                }
              >
                Previous
              </Button>
              <Button
                variant="outline"
                disabled={items.length < pageSize || calls.isFetching}
                onClick={() => setOffset((value) => value + pageSize)}
              >
                Next
              </Button>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
