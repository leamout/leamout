"use client";

import { Button } from "@leamout/ui/components/button";
import Link from "next/link";
import { useQueryClient } from "@tanstack/react-query";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { useAgents, useTrunks, useIntegrations, useAllTools, resourceKeys } from "@/lib/resources";
import { RequestError, RequestLoading } from "./request-state";

export function ResourceList({ kind }: { kind: "agents" | "trunks" | "integrations" | "tools" }) {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const queryClient = useQueryClient();
  const agents = useAgents();
  const trunks = useTrunks();
  const integrations = useIntegrations();
  const tools = useAllTools();
  const query = kind === "agents" ? agents : kind === "trunks" ? trunks : kind === "integrations" ? integrations : tools;
  const data = kind === "agents" ? agents.data : kind === "trunks" ? trunks.data : kind === "integrations" ? integrations.data : tools.data;
  const title = { agents: "Agents", trunks: "SIP trunks", integrations: "AI integrations", tools: "Tools" }[kind];
  const createHref = { agents: "/agents/new", trunks: "/sip-trunks/new", integrations: "/ai-providers", tools: "/tools/new" }[kind];
  const key = kind === "agents" ? resourceKeys.agents(organizationId)
    : kind === "trunks" ? resourceKeys.trunks(organizationId)
    : kind === "integrations" ? resourceKeys.integrations(organizationId) : resourceKeys.tools(organizationId);
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between gap-4">
        <h1 className="text-3xl font-semibold">{title}</h1>
        {kind !== "integrations" && <Button render={<Link href={createHref} />}>Create {kind === "agents" ? "agent" : kind === "trunks" ? "trunk" : "tool"}</Button>}
      </div>
      {kind === "integrations" && <p className="text-sm text-muted-foreground">Select a provider to add or manage organization-owned credentials.</p>}
      {query.isPending ? <RequestLoading what={title.toLowerCase()} /> : query.isError ? (
        <RequestError error={query.error} retry={() => void query.refetch()} />
      ) : kind === "integrations" ? (
        <div className="space-y-2">
          {["openai", "gemini", "deepgram", "assemblyai", "groq", "cartesia", "elevenlabs"].map(provider => {
            const items = integrations.data?.filter(item => item.provider === provider) ?? [];
            return <Link key={provider} href={`/ai-providers/${provider}`} className="flex justify-between rounded-lg border p-4 hover:bg-muted">
              <span className="capitalize font-medium">{provider}</span><span className="text-sm text-muted-foreground">{items.length} credentials</span>
            </Link>;
          })}
        </div>
      ) : !data?.length ? (
        <p className="rounded-lg border p-6 text-sm text-muted-foreground">No {title.toLowerCase()} yet.</p>
      ) : (
        <div className="divide-y rounded-lg border">
          {kind === "agents" && agents.data?.map(item => <Link key={item.id} href={`/agents/${item.id}`} className="flex justify-between gap-4 p-4 hover:bg-muted"><span className="font-medium">{item.name}</span><span className="text-sm text-muted-foreground">{item.engine} · {item.status}</span></Link>)}
          {kind === "trunks" && trunks.data?.map(item => <Link key={item.id} href={`/sip-trunks/${item.id}`} className="flex justify-between gap-4 p-4 hover:bg-muted"><span className="font-medium">{item.name}</span><span className="text-sm text-muted-foreground">{item.direction} · {item.status}</span></Link>)}
          {kind === "tools" && tools.data?.map(item => <Link key={item.id} href={`/tools/${item.id}`} className="flex justify-between gap-4 p-4 hover:bg-muted"><span className="font-medium">{item.name}</span><span className="text-sm text-muted-foreground">{item.agent_name} · {item.type}</span></Link>)}
        </div>
      )}
      <Button variant="outline" onClick={() => void queryClient.invalidateQueries({queryKey:key})}>Refresh</Button>
    </div>
  );
}
