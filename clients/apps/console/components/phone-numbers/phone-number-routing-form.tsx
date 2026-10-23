"use client";

import { Button } from "@leamout/ui/components/button";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useState } from "react";
import type { VoiceAgent } from "@/components/agents/types";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { api, errorMessage } from "@/lib/api";
import type { AgentNumberBinding, PhoneNumber } from "./types";

export function PhoneNumberRoutingForm({ number }: { number: PhoneNumber }) {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const client = useQueryClient();
  const [selectedAgentId, setSelectedAgentId] = useState("");
  const [confirmUnbind, setConfirmUnbind] = useState(false);

  const agents = useQuery({
    queryKey: ["voice-agents", organizationId],
    queryFn: ({ signal }) =>
      api<{ voice_agents: VoiceAgent[] }>("/voice-agents/", {
        organizationId,
        signal,
      }),
    enabled: Boolean(organizationId),
  });

  // The Go API lists bindings by agent, not by number.
  const bindings = useQuery({
    queryKey: ["number-agent-bindings", organizationId, number.id],
    queryFn: async ({ signal }) => {
      const all = await Promise.all(
        (agents.data?.voice_agents ?? []).map((agent) =>
          api<{ bindings: AgentNumberBinding[] }>(
            `/voice-agents/${encodeURIComponent(agent.id)}/bindings`,
            {
              organizationId,
              signal,
            },
          ),
        ),
      );

      return (
        all
          .flatMap((result) => result.bindings)
          .find((binding) => binding.phone_number_id === number.id) ?? null
      );
    },
    enabled: Boolean(organizationId) && agents.isSuccess,
  });

  const bind = useMutation({
    mutationFn: () =>
      api<AgentNumberBinding>(
        `/voice-agents/${encodeURIComponent(selectedAgentId)}/bindings`,
        {
          method: "POST",
          organizationId,
          body: { phone_number_id: number.id },
        },
      ),
    onSuccess: async () => {
      await client.invalidateQueries({
        queryKey: ["number-agent-bindings", organizationId, number.id],
      });
      await client.invalidateQueries({
        queryKey: ["agent-number-bindings", organizationId],
      });
    },
  });

  const unbind = useMutation({
    mutationFn: (binding: AgentNumberBinding) =>
      api<void>(
        "/voice-agents/" +
          encodeURIComponent(binding.voice_agent_id) +
          "/bindings/" +
          encodeURIComponent(binding.id),
        {
          method: "DELETE",
          organizationId,
        },
      ),
    onSuccess: async () => {
      setConfirmUnbind(false);
      await client.invalidateQueries({
        queryKey: ["number-agent-bindings", organizationId, number.id],
      });
      await client.invalidateQueries({
        queryKey: ["agent-number-bindings", organizationId],
      });
    },
  });

  const assigned = bindings.data;
  const assignedName = agents.data?.voice_agents.find(
    (agent) => agent.id === assigned?.voice_agent_id,
  )?.name;

  return (
    <div className="max-w-2xl space-y-5">
      <h2 className="text-lg font-semibold">Inbound Agent routing</h2>
      {agents.isPending || bindings.isPending ? (
        <p role="status">Loading routing…</p>
      ) : agents.isError || bindings.isError ? (
        <div role="alert">
          <p>{errorMessage(agents.error ?? bindings.error)}</p>
          <Button
            variant="outline"
            onClick={() => {
              void agents.refetch();
              void bindings.refetch();
            }}
          >
            Retry
          </Button>
        </div>
      ) : assigned ? (
        <div className="space-y-3 rounded-lg border p-5">
          <p>
            Assigned Agent:{" "}
            <Link
              className="font-medium underline"
              href={`/agents/${encodeURIComponent(assigned.voice_agent_id)}`}
            >
              {assignedName ?? assigned.voice_agent_id}
            </Link>
          </p>
          {!confirmUnbind ? (
            <Button variant="outline" onClick={() => setConfirmUnbind(true)}>
              Unbind agent…
            </Button>
          ) : (
            <div className="flex gap-2">
              <Button
                variant="destructive"
                disabled={unbind.isPending}
                onClick={() => unbind.mutate(assigned)}
              >
                Confirm unbind
              </Button>
              <Button variant="outline" onClick={() => setConfirmUnbind(false)}>
                Cancel
              </Button>
            </div>
          )}
          {unbind.isError && (
            <p role="alert" className="text-destructive">
              {errorMessage(unbind.error)}
            </p>
          )}
        </div>
      ) : (
        <div className="space-y-4 rounded-lg border p-5">
          <label htmlFor="assigned-agent" className="block text-sm font-medium">
            Route calls to Agent
          </label>
          <select
            id="assigned-agent"
            className="h-10 w-full rounded-md border bg-background px-3"
            value={selectedAgentId}
            onChange={(event) => setSelectedAgentId(event.target.value)}
          >
            <option value="">Choose Agent</option>
            {agents.data?.voice_agents.map((agent) => (
              <option key={agent.id} value={agent.id}>
                {agent.name}
              </option>
            ))}
          </select>
          <p className="text-xs text-muted-foreground">
            A number may only be bound to one Agent. Unbind an existing
            assignment before switching.
          </p>
          <Button
            disabled={
              !selectedAgentId || bind.isPending || !number.voice_enabled
            }
            onClick={() => bind.mutate()}
          >
            {bind.isPending ? "Assigning…" : "Assign Agent"}
          </Button>
          {!number.voice_enabled && (
            <p className="text-sm text-muted-foreground">
              Enable voice in Configuration first.
            </p>
          )}
          {bind.isError && (
            <p role="alert" className="text-destructive">
              {errorMessage(bind.error)}
            </p>
          )}
        </div>
      )}
    </div>
  );
}
