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
import type { VoiceAgent } from "./types";

function formattedDate(value: string): string {
  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? "—"
    : date.toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
      });
}

export function AgentList() {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const [search, setSearch] = useState("");
  const [engine, setEngine] = useState("all");
  const query = useQuery({
    queryKey: ["voice-agents", organizationId],
    queryFn: ({ signal }) =>
      api<{ voice_agents: VoiceAgent[] }>("/voice-agents/", {
        organizationId,
        signal,
      }),
    enabled: organizationId !== "",
  });
  const agents = query.data?.voice_agents ?? [];
  const filtered = agents.filter(
    (agent) =>
      agent.name.toLowerCase().includes(search.trim().toLowerCase()) &&
      (engine === "all" || agent.engine === engine),
  );

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="space-y-2">
          <h1 className="text-3xl font-semibold tracking-tight">Agents</h1>
          <p className="text-muted-foreground">
            Create and configure autonomous voice agents in your organization.
          </p>
        </div>
        <Button render={<Link href="/agents/new" />}>Create agent</Button>
      </div>
      <div className="flex flex-wrap gap-4">
        <div className="flex-1 space-y-2">
          <label htmlFor="agent-search" className="text-sm font-medium">
            Search agents
          </label>
          <Input
            id="agent-search"
            type="search"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Search by name"
          />
        </div>
        <div className="space-y-2">
          <label htmlFor="agent-engine" className="block text-sm font-medium">
            Engine
          </label>
          <select
            id="agent-engine"
            className="h-9 rounded-md border bg-background px-3 text-sm"
            value={engine}
            onChange={(event) => setEngine(event.target.value)}
          >
            <option value="all">All engines</option>
            <option value="realtime">Realtime</option>
            <option value="composable">Composable</option>
          </select>
        </div>
      </div>
      {query.isError ? (
        <div className="space-y-3 rounded-lg border p-6" role="alert">
          <p>{errorMessage(query.error)}</p>
          <Button variant="outline" onClick={() => void query.refetch()}>
            Try again
          </Button>
        </div>
      ) : query.isPending ? (
        <p role="status" className="rounded-lg border p-6 text-muted-foreground">
          Loading agents…
        </p>
      ) : (
        <div className="overflow-hidden rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Engine</TableHead>
                <TableHead>Language</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Updated</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filtered.map((agent) => (
                <TableRow key={agent.id}>
                  <TableCell>
                    <Link
                      className="font-medium underline-offset-4 hover:underline"
                      href={"/agents/" + encodeURIComponent(agent.id)}
                    >
                      {agent.name}
                    </Link>
                  </TableCell>
                  <TableCell>
                    {agent.engine === "composable" ? "STT → LLM → TTS" : "Realtime"}
                  </TableCell>
                  <TableCell>{agent.language}</TableCell>
                  <TableCell>{agent.status}</TableCell>
                  <TableCell>{formattedDate(agent.updated_at)}</TableCell>
                </TableRow>
              ))}
              {filtered.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className="h-40 text-center">
                    <p className="font-medium">
                      {agents.length === 0
                        ? "No agents yet"
                        : "No matching agents"}
                    </p>
                    <p className="mt-2 text-muted-foreground">
                      {agents.length === 0
                        ? "Create your first Voice Agent to get started."
                        : "Try another name or engine filter."}
                    </p>
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
