"use client";

import { Button } from "@leamout/ui/components/button";
import { Input } from "@leamout/ui/components/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@leamout/ui/components/table";
import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { useState } from "react";
import { useOrganization } from "@/components/onboarding/organization-provider";
import type { VoiceAgent } from "@/components/agents/types";
import { api, errorMessage } from "@/lib/api";
import type { Tool } from "./types";

export function ToolList() {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const [search, setSearch] = useState("");
  const [type, setType] = useState("all");
  const agents = useQuery({
    queryKey: ["voice-agents", organizationId],
    queryFn: ({ signal }) => api<{ voice_agents: VoiceAgent[] }>(
      "/voice-agents/", { organizationId, signal },
    ),
    enabled: Boolean(organizationId),
  });
  const tools = useQuery({
    queryKey: ["tools", organizationId, agents.data?.voice_agents.map((item) => item.id)],
    queryFn: async ({ signal }) => {
      const rows = await Promise.all(
        (agents.data?.voice_agents ?? []).map((agent) =>
          api<{ tools: Tool[] }>(
            "/voice-agents/" + encodeURIComponent(agent.id) + "/tools/",
            { organizationId, signal },
          ),
        ),
      );
      return rows.flatMap((result) => result.tools);
    },
    enabled: Boolean(organizationId) && agents.isSuccess,
  });
  const agentNames = new Map(
    agents.data?.voice_agents.map((agent) => [agent.id, agent.name] as const) ?? [],
  );
  const items = tools.data ?? [];
  const filtered = items.filter((item) =>
    (item.name + " " + item.description + " " + (agentNames.get(item.voice_agent_id) ?? ""))
      .toLowerCase().includes(search.trim().toLowerCase()) &&
    (type === "all" || item.type === type),
  );

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div className="space-y-2">
          <h1 className="text-3xl font-semibold">Tools</h1>
          <p className="text-muted-foreground">Manage tools attached to Voice Agents.</p>
        </div>
        <Button render={<Link href="/tools/new" />}>Create tool</Button>
      </div>
      <div className="flex gap-4">
        <label className="flex-1 space-y-2 text-sm">
          <span className="block font-medium">Search tools</span>
          <Input type="search" value={search} onChange={(event) => setSearch(event.target.value)} />
        </label>
        <label className="space-y-2 text-sm">
          <span className="block font-medium">Type</span>
          <select className="block h-9 rounded-md border bg-background px-3"
            value={type} onChange={(event) => setType(event.target.value)}>
            <option value="all">All</option>
            <option value="builtin">Built-in</option>
            <option value="webhook">Webhook</option>
          </select>
        </label>
      </div>
      {agents.isPending || tools.isPending ? (
        <p role="status">Loading Agent tools…</p>
      ) : agents.isError || tools.isError ? (
        <div role="alert" className="space-y-3">
          <p>{errorMessage(agents.error ?? tools.error)}</p>
          <Button variant="outline" onClick={() => {
            void agents.refetch();
            void tools.refetch();
          }}>Retry</Button>
        </div>
      ) : (
        <div className="overflow-x-auto rounded-lg border">
          <Table>
            <TableHeader><TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Agent</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Timeout</TableHead>
            </TableRow></TableHeader>
            <TableBody>
              {filtered.map((tool) => (
                <TableRow key={tool.id}>
                  <TableCell>
                    <Link className="font-medium hover:underline" href={"/tools/" + encodeURIComponent(tool.id) + "?agent_id=" + encodeURIComponent(tool.voice_agent_id)}>{tool.name}</Link>
                    <p className="text-xs text-muted-foreground">{tool.description}</p>
                  </TableCell>
                  <TableCell>{tool.type}</TableCell>
                  <TableCell>{agentNames.get(tool.voice_agent_id) ?? tool.voice_agent_id}</TableCell>
                  <TableCell>{tool.enabled ? "Enabled" : "Disabled"}</TableCell>
                  <TableCell>{tool.timeout_ms} ms</TableCell>
                </TableRow>
              ))}
              {!filtered.length && <TableRow><TableCell colSpan={5} className="py-12 text-center">No matching Agent tools.</TableCell></TableRow>}
            </TableBody>
          </Table>
        </div>
      )}
      <p className="text-xs text-muted-foreground">Tools are fetched per active Agent. Disabled Agents are not included in this list.</p>
    </div>
  );
}
