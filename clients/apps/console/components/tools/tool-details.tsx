"use client";

import { Button } from "@leamout/ui/components/button";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { SecretReveal } from "@/components/webhooks/secret-reveal";
import { api, errorMessage } from "@/lib/api";
import { ToolForm } from "./tool-form";
import type { Tool } from "./types";

export function ToolDetails({
  toolId,
  agentId,
}: {
  toolId: string;
  agentId: string;
}) {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const client = useQueryClient();
  const router = useRouter();
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [confirmRotate, setConfirmRotate] = useState(false);
  const base = `/voice-agents/${encodeURIComponent(agentId)}/tools/`;

  const tools = useQuery({
    queryKey: ["agent-tools", organizationId, agentId],
    queryFn: ({ signal }) =>
      api<{ tools: Tool[] }>(base, {
        organizationId,
        signal,
      }),
    enabled: Boolean(organizationId) && Boolean(agentId),
  });

  const tool = tools.data?.tools.find((item) => item.id === toolId);

  const rotate = useMutation({
    mutationFn: () =>
      api<{ signing_secret: string }>(
        `${base + encodeURIComponent(toolId)}/rotate-signing-secret`,
        {
          method: "POST",
          organizationId,
        },
      ),
  });

  const remove = useMutation({
    mutationFn: () =>
      api<void>(base + encodeURIComponent(toolId), {
        method: "DELETE",
        organizationId,
      }),
    onSuccess: async () => {
      await client.invalidateQueries({
        queryKey: ["tools", organizationId],
      });
      await client.invalidateQueries({
        queryKey: ["agent-tools", organizationId, agentId],
      });
      router.replace("/tools");
    },
  });

  if (!agentId) {
    return (
      <div className="space-y-3">
        <h1 className="text-2xl font-semibold">Agent context required</h1>
        <p className="text-muted-foreground">
          Tool detail URLs require an agent_id query parameter because the Go
          API exposes tools only within an Agent.
        </p>
        <Link href="/tools" className="underline">
          Back to tools
        </Link>
      </div>
    );
  }

  if (tools.isPending) {
    return <p role="status">Loading tool…</p>;
  }

  if (tools.isError) {
    return (
      <div role="alert" className="space-y-3">
        <p>{errorMessage(tools.error)}</p>
        <Button variant="outline" onClick={() => void tools.refetch()}>
          Retry
        </Button>
      </div>
    );
  }

  if (!tool) {
    return (
      <p role="alert">
        Tool not found in this Agent.{" "}
        <Link href="/tools" className="underline">
          Back to tools
        </Link>
      </p>
    );
  }

  return (
    <div className="space-y-7">
      <Link
        href="/tools"
        className="text-sm text-muted-foreground hover:underline"
      >
        Back to tools
      </Link>

      <div className="space-y-2">
        <h1 className="text-3xl font-semibold">{tool.name}</h1>
        <p className="break-all font-mono text-xs text-muted-foreground">
          {tool.id}
        </p>
        <p className="text-sm text-muted-foreground">
          {tool.type} · {tool.enabled ? "Enabled" : "Disabled"}
        </p>
        <Link
          className="text-sm underline"
          href={`/agents/${encodeURIComponent(agentId)}`}
        >
          View Agent
        </Link>
      </div>

      <ToolForm tool={tool} key={tool.id} />

      {tool.type === "webhook" && (
        <section className="space-y-4 rounded-lg border p-5">
          <h2 className="text-lg font-semibold">Tool signing secret</h2>
          <p className="text-sm text-muted-foreground">
            Rotation replaces the secret used to sign tool execution requests.
          </p>
          {rotate.data && (
            <SecretReveal
              secret={rotate.data.signing_secret}
              label="Replacement tool secret"
            />
          )}
          {!confirmRotate ? (
            <Button variant="outline" onClick={() => setConfirmRotate(true)}>
              Rotate secret…
            </Button>
          ) : (
            <div className="flex gap-2">
              <Button
                variant="destructive"
                disabled={rotate.isPending}
                onClick={() => rotate.mutate()}
              >
                Confirm rotation
              </Button>
              <Button variant="outline" onClick={() => setConfirmRotate(false)}>
                Cancel
              </Button>
            </div>
          )}
          {rotate.isError && <p role="alert">{errorMessage(rotate.error)}</p>}
        </section>
      )}

      <section className="space-y-3 border-t pt-6">
        <h2 className="font-semibold">Remove tool</h2>
        {!confirmDelete ? (
          <Button variant="outline" onClick={() => setConfirmDelete(true)}>
            Remove tool…
          </Button>
        ) : (
          <div className="flex gap-2">
            <Button
              variant="destructive"
              disabled={remove.isPending}
              onClick={() => remove.mutate()}
            >
              Confirm removal
            </Button>
            <Button variant="outline" onClick={() => setConfirmDelete(false)}>
              Cancel
            </Button>
          </div>
        )}
        {remove.isError && <p role="alert">{errorMessage(remove.error)}</p>}
      </section>
    </div>
  );
}
