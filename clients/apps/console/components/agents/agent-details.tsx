"use client";

import { Button } from "@leamout/ui/components/button";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@leamout/ui/components/tabs";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { ApiError, api, errorMessage } from "@/lib/api";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { AgentForm } from "@/components/agents/agent-form";
import { AgentProviderSettings } from "@/components/agents/agent-provider-settings";
import { AgentReadiness } from "@/components/agents/agent-readiness";
import { AgentRouting } from "@/components/agents/agent-routing";
import { AgentTools } from "@/components/agents/agent-tools";
import type { VoiceAgent } from "./types";

const planned = [
  {
    value: "test",
    title: "Test",
    description: "Interactive agent testing is not connected yet.",
  },
  {
    value: "activity",
    title: "Activity",
    description: "Call history and diagnostics will appear in Calls when connected.",
    href: "/calls",
    link: "View calls",
  },
];

export function AgentDetails({ agentId }: { agentId: string }) {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const client = useQueryClient();
  const router = useRouter();
  const [confirmDisable, setConfirmDisable] = useState(false);

  const query = useQuery({
    queryKey: ["voice-agents", organizationId, agentId],
    queryFn: ({ signal }) =>
      api<VoiceAgent>("/voice-agents/" + encodeURIComponent(agentId), {
        organizationId,
        signal,
      }),
    enabled: organizationId !== "" && agentId !== "",
  });

  const disable = useMutation({
    mutationFn: () =>
      api<void>("/voice-agents/" + encodeURIComponent(agentId), {
        method: "DELETE",
        organizationId,
      }),
    onSuccess: async () => {
      client.removeQueries({
        queryKey: ["voice-agents", organizationId, agentId],
        exact: true,
      });
      await client.invalidateQueries({
        queryKey: ["voice-agents", organizationId],
      });
      router.replace("/agents");
    },
  });

  if (query.isError) {
    return (
      <div className="space-y-4" role="alert">
        <h1 className="text-2xl font-semibold">
          {query.error instanceof ApiError && query.error.status === 404
            ? "Agent not found"
            : "Unable to load agent"}
        </h1>
        <p className="text-muted-foreground">{errorMessage(query.error)}</p>
        <Button variant="outline" onClick={() => void query.refetch()}>
          Try again
        </Button>
        <Link href="/agents" className="block text-sm underline underline-offset-4">
          Back to agents
        </Link>
      </div>
    );
  }

  if (!query.data) {
    return (
      <p role="status" className="text-sm text-muted-foreground">
        Loading agent…
      </p>
    );
  }

  const agent = query.data;
  return (
    <div className="space-y-8">
      <Link
        href="/agents"
        className="text-sm text-muted-foreground hover:underline"
      >
        Back to agents
      </Link>
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">{agent.name}</h1>
        <p className="break-all font-mono text-xs text-muted-foreground">
          Agent ID: {agent.id}
        </p>
        <p className="text-sm text-muted-foreground">
          {agent.engine === "realtime" ? "Realtime" : "Composable"} ·{" "}
          {agent.language} · Revision {agent.configuration_revision}
        </p>
      </div>

      <Tabs defaultValue="configuration" className="gap-6">
        <div className="overflow-x-auto">
          <TabsList aria-label="Agent settings">
            <TabsTrigger value="configuration">Configuration</TabsTrigger>
            <TabsTrigger value="readiness">Readiness</TabsTrigger>
            <TabsTrigger value="providers">Providers</TabsTrigger>
            <TabsTrigger value="routing">Routing</TabsTrigger>
            <TabsTrigger value="tools">Tools</TabsTrigger>
            {planned.map((section) => (
              <TabsTrigger key={section.value} value={section.value}>
                {section.title}
              </TabsTrigger>
            ))}
          </TabsList>
        </div>
        <TabsContent value="configuration">
          <AgentForm key={agent.id} agent={agent} />
        </TabsContent>
        <TabsContent value="readiness">
          <AgentReadiness agentId={agent.id} />
        </TabsContent>
        <TabsContent value="providers">
          <AgentProviderSettings agent={agent} />
        </TabsContent>
        <TabsContent value="routing">
          <AgentRouting agentId={agent.id} />
        </TabsContent>
        <TabsContent value="tools">
          <AgentTools agentId={agent.id} />
        </TabsContent>
        {planned.map((section) => (
          <TabsContent key={section.value} value={section.value}>
            <div className="space-y-4 rounded-lg border p-6">
              <h2 className="text-lg font-semibold">{section.title}</h2>
              <p className="text-muted-foreground">{section.description}</p>
              {section.href && (
                <Button
                  variant="outline"
                  render={<Link href={section.href} />}
                >
                  {section.link}
                </Button>
              )}
            </div>
          </TabsContent>
        ))}
      </Tabs>

      <section className="space-y-3 border-t pt-8" aria-labelledby="disable-title">
        <h2 id="disable-title" className="text-lg font-semibold">
          Disable agent
        </h2>
        <p className="max-w-2xl text-sm text-muted-foreground">
          Disabling removes this agent from the active list. It cannot be
          reactivated through the current API and still counts toward your
          plan's Voice Agent limit.
        </p>
        {!confirmDisable ? (
          <Button variant="outline" onClick={() => setConfirmDisable(true)}>
            Disable agent…
          </Button>
        ) : (
          <div className="space-y-3 rounded-lg border p-4">
            <p className="text-sm font-medium">
              Disable this agent? This action cannot be undone in the Console.
            </p>
            <div className="flex flex-wrap gap-2">
              <Button
                variant="destructive"
                disabled={disable.isPending}
                onClick={() => disable.mutate()}
              >
                {disable.isPending ? "Disabling…" : "Confirm disable"}
              </Button>
              <Button
                variant="outline"
                disabled={disable.isPending}
                onClick={() => setConfirmDisable(false)}
              >
                Cancel
              </Button>
            </div>
          </div>
        )}
        {disable.isError && (
          <p role="alert" className="text-sm text-destructive">
            {errorMessage(disable.error)}
          </p>
        )}
      </section>
    </div>
  );
}
