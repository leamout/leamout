import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { useOrganization } from "@/components/onboarding/organization-provider";

export type VoiceAgent = {
  id: string;
  organization_id: string;
  name: string;
  engine: "realtime" | "composable";
  instructions: string;
  voice?: string | null;
  language: string;
  status: string;
  configuration_revision: number;
  updated_at: string;
};

export type SipTrunk = {
  id: string;
  organization_id: string;
  name: string;
  direction: "inbound" | "outbound" | "bidirectional";
  status: "active" | "disabled";
  max_cps: number;
  max_concurrent_calls: number;
  codecs: string[];
  inbound_enabled: boolean;
  updated_at: string;
};

export type TrunkEndpoint = {
  id: string;
  host: string;
  port: number;
  transport: string;
  direction: string;
  enabled: boolean;
};

export type AiIntegration = {
  id: string;
  organization_id: string;
  provider: string;
  name: string;
  connection_state: string;
  verified_at?: string | null;
  failure_code?: string | null;
  capabilities: string[];
  voice_agent_ids: string[];
};

export type AgentTool = {
  id: string;
  organization_id: string;
  voice_agent_id: string;
  name: string;
  description: string;
  type: "builtin" | "webhook";
  parameters: Record<string, unknown>;
  endpoint_url?: string | null;
  timeout_ms: number;
  enabled: boolean;
  signing_secret?: string;
};

export type AgentToolWithOwner = AgentTool & { agent_name: string };

export const resourceKeys = {
  agents: (id: string) => ["resources", "agents", id] as const,
  trunks: (id: string) => ["resources", "trunks", id] as const,
  integrations: (id: string) => ["resources", "integrations", id] as const,
  tools: (id: string) => ["resources", "tools", id] as const,
  agentTools: (id: string, agentId: string) =>
    ["resources", "tools", id, agentId] as const,
};

export async function listAgents(organizationId: string, signal?: AbortSignal) {
  const data = await api<{ voice_agents: VoiceAgent[] }>("/voice-agents", {
    organizationId,
    signal,
  });
  return data.voice_agents;
}

export async function listTrunks(organizationId: string, signal?: AbortSignal) {
  const data = await api<{ trunks: SipTrunk[] }>("/trunks", {
    organizationId,
    signal,
  });
  return data.trunks;
}

export async function listIntegrations(
  organizationId: string,
  signal?: AbortSignal,
) {
  const data = await api<{ ai_integrations: AiIntegration[] }>(
    "/ai-integrations",
    { organizationId, signal },
  );
  return data.ai_integrations;
}

export async function listAgentTools(
  organizationId: string,
  agentId: string,
  signal?: AbortSignal,
) {
  const data = await api<{ tools: AgentTool[] }>(
    `/voice-agents/${encodeURIComponent(agentId)}/tools`,
    { organizationId, signal },
  );
  return data.tools;
}

export async function listAllTools(
  organizationId: string,
  signal?: AbortSignal,
) {
  const agents = await listAgents(organizationId, signal);
  const perAgent = await Promise.all(
    agents.map(async (agent) => {
      const tools = await listAgentTools(organizationId, agent.id, signal);
      return tools.map((tool): AgentToolWithOwner => ({
        ...tool,
        agent_name: agent.name,
      }));
    }),
  );
  return perAgent.flat();
}

export function useAgents() {
  const { selected } = useOrganization();
  const id = selected?.id ?? "";
  return useQuery({
    queryKey: resourceKeys.agents(id),
    queryFn: ({ signal }) => listAgents(id, signal),
    enabled: Boolean(id),
  });
}

export function useTrunks() {
  const { selected } = useOrganization();
  const id = selected?.id ?? "";
  return useQuery({
    queryKey: resourceKeys.trunks(id),
    queryFn: ({ signal }) => listTrunks(id, signal),
    enabled: Boolean(id),
  });
}

export function useIntegrations() {
  const { selected } = useOrganization();
  const id = selected?.id ?? "";
  return useQuery({
    queryKey: resourceKeys.integrations(id),
    queryFn: ({ signal }) => listIntegrations(id, signal),
    enabled: Boolean(id),
  });
}

export function useTools(agentId: string) {
  const { selected } = useOrganization();
  const id = selected?.id ?? "";
  return useQuery({
    queryKey: resourceKeys.agentTools(id, agentId),
    queryFn: ({ signal }) => listAgentTools(id, agentId, signal),
    enabled: Boolean(id) && Boolean(agentId),
  });
}

export function useAllTools() {
  const { selected } = useOrganization();
  const id = selected?.id ?? "";
  return useQuery({
    queryKey: resourceKeys.tools(id),
    queryFn: ({ signal }) => listAllTools(id, signal),
    enabled: Boolean(id),
  });
}
