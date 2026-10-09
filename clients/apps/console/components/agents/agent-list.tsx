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

export type AgentListItem = {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
  ready: boolean;
  configuration: string;
  updated: string;
};

export function AgentList({ agents = [] }: { agents?: AgentListItem[] }) {
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState("all");
  const filtered = agents.filter(
    (agent) =>
      `${agent.name} ${agent.description}`
        .toLowerCase()
        .includes(search.toLowerCase()) &&
      (status === "all" ||
        (status === "enabled" ? agent.enabled : !agent.enabled)),
  );

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="space-y-2">
          <h1 className="text-3xl font-semibold tracking-tight">Agents</h1>
          <p className="text-muted-foreground">
            Configure your autonomous voice agents.
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
            placeholder="Search by name or description"
          />
        </div>
        <div className="space-y-2">
          <label htmlFor="agent-status" className="block text-sm font-medium">
            Status
          </label>
          <select
            id="agent-status"
            className="h-9 rounded-md border bg-background px-3 text-sm"
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
              <TableHead>Name</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>AI configuration</TableHead>
              <TableHead>Updated</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filtered.map((agent) => (
              <TableRow key={agent.id}>
                <TableCell>
                  <Link
                    className="font-medium underline-offset-4 hover:underline"
                    href={`/agents/${encodeURIComponent(agent.id)}`}
                  >
                    {agent.name}
                  </Link>
                  <p className="text-sm text-muted-foreground">
                    {agent.description}
                  </p>
                </TableCell>
                <TableCell>
                  {agent.enabled ? "Enabled" : "Disabled"}
                  <p className="text-xs text-muted-foreground">
                    {agent.ready ? "Ready" : "Setup required"}
                  </p>
                </TableCell>
                <TableCell>{agent.configuration}</TableCell>
                <TableCell>{agent.updated}</TableCell>
              </TableRow>
            ))}
            {filtered.length === 0 && (
              <TableRow>
                <TableCell colSpan={4} className="h-40 text-center">
                  <p className="font-medium">
                    {agents.length === 0
                      ? "No agents to display"
                      : "No matching agents"}
                  </p>
                  <p className="mt-2 text-muted-foreground">
                    {agents.length === 0
                      ? "Your agents will appear here when account integration is connected."
                      : "Try another search or status filter."}
                  </p>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <p className="text-sm text-muted-foreground">
        Enabled status and call readiness are separate. This page is a UI
        preview.
      </p>
    </div>
  );
}
