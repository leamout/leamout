import { api } from "@/lib/api";
import type { VoiceAgent, VoiceAgentInput } from "./model";

const path = "/voice-agents";

export const voiceAgentKeys = {
  list: (organizationId: string) => ["voice-agents", organizationId] as const,
  detail: (organizationId: string, agentId: string) =>
    ["voice-agents", organizationId, agentId] as const,
};

export function listVoiceAgents(organizationId: string, signal?: AbortSignal) {
  return api<{ voice_agents: VoiceAgent[] }>(path + "/", {
    organizationId,
    signal,
  });
}

export function getVoiceAgent(
  organizationId: string,
  agentId: string,
  signal?: AbortSignal,
) {
  return api<VoiceAgent>(path + "/" + encodeURIComponent(agentId), {
    organizationId,
    signal,
  });
}

export function createVoiceAgent(
  organizationId: string,
  input: VoiceAgentInput,
  idempotencyKey: string,
) {
  return api<VoiceAgent>(path + "/", {
    method: "POST",
    organizationId,
    idempotencyKey,
    body: input,
  });
}

export function updateVoiceAgent(
  organizationId: string,
  agentId: string,
  input: Partial<VoiceAgentInput>,
) {
  return api<VoiceAgent>(path + "/" + encodeURIComponent(agentId), {
    method: "PATCH",
    organizationId,
    body: input,
  });
}

export function disableVoiceAgent(organizationId: string, agentId: string) {
  return api<void>(path + "/" + encodeURIComponent(agentId), {
    method: "DELETE",
    organizationId,
  });
}
