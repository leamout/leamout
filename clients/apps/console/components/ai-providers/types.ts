export type ProviderRole = "realtime" | "stt" | "llm" | "tts";

export type AIIntegration = {
  id: string;
  organization_id: string;
  provider: string;
  name: string;
  connection_state: "unchecked" | "ready" | "invalid" | "unavailable";
  verified_at?: string | null;
  failure_code?: string | null;
  capabilities: string[];
  voice_agent_ids: string[];
  created_at: string;
  rotated_at: string;
  updated_at: string;
};

export type AgentProviderBinding = {
  id: string;
  organization_id: string;
  voice_agent_id: string;
  role: ProviderRole;
  provider: string;
  credential_id?: string | null;
  config: Record<string, unknown>;
};

export type AgentReadinessReport = {
  ready: boolean;
  configuration_revision: number;
  engine: string;
  bindings: {
    role: ProviderRole;
    provider: string;
    integration_id?: string | null;
    connection_state: string;
    failure_code?: string | null;
    config: Record<string, unknown>;
  }[];
  issues: {
    code: string;
    field: string;
    message: string;
    remediation: string;
  }[];
};
