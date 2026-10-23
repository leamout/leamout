"use client";

import { Button } from "@leamout/ui/components/button";
import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { useOrganization } from "@/components/onboarding/organization-provider";
import type { Tool } from "@/components/tools/types";
import { api, errorMessage } from "@/lib/api";

export function AgentTools({ agentId }: { agentId: string }) {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";

  const tools = useQuery({
    queryKey: ["agent-tools", organizationId, agentId],
    queryFn: ({ signal }) =>
      api<{ tools: Tool[] }>(
        "/voice-agents/" + encodeURIComponent(agentId) + "/tools/",
        {
          organizationId,
          signal,
        },
      ),
    enabled: Boolean(organizationId),
  });

  return (
    <section className="space-y-5">
      <div className="space-y-2">
        <h2 className="text-lg font-semibold">Agent tools</h2>
        <p className="text-sm text-muted-foreground">
          Built-in actions and signed webhook tool calls available to this
          Agent.
        </p>
        <Link
          className="text-sm underline"
          href={"/tools/new?agent_id=" + encodeURIComponent(agentId)}
        >
          Create tool for this Agent
        </Link>
      </div>

      {tools.isPending ? (
        <p role="status">Loading Agent tools…</p>
      ) : tools.isError ? (
        <div role="alert" className="space-y-3">
          <p>{errorMessage(tools.error)}</p>
          <Button variant="outline" onClick={() => void tools.refetch()}>
            Retry
          </Button>
        </div>
      ) : tools.data.tools.length ? (
        <ul className="space-y-3">
          {tools.data.tools.map((tool) => (
            <li key={tool.id} className="space-y-2 rounded-lg border p-4">
              <Link
                className="font-medium hover:underline"
                href={
                  "/tools/" +
                  encodeURIComponent(tool.id) +
                  "?agent_id=" +
                  encodeURIComponent(agentId)
                }
              >
                {tool.name}
              </Link>
              <p className="text-sm text-muted-foreground">
                {tool.description}
              </p>
              <p className="text-sm">
                {tool.type} · {tool.enabled ? "Enabled" : "Disabled"} ·{" "}
                {tool.timeout_ms} ms
              </p>
            </li>
          ))}
        </ul>
      ) : (
        <p className="rounded-lg border border-dashed p-6 text-sm text-muted-foreground">
          No tools attached to this Agent.
        </p>
      )}
    </section>
  );
}
