export type Tool = {
  id: string;
  organization_id: string;
  voice_agent_id: string;
  type: "builtin" | "webhook";
  name: string;
  description: string;
  parameters: Record<string, unknown>;
  endpoint_url?: string | null;
  timeout_ms: number;
  enabled: boolean;
  created_at: string;
  updated_at: string;
  signing_secret?: string;
};

export const builtins = [
  "hangup_call",
  "transfer_call",
  "hold_call",
  "resume_call",
  "send_dtmf",
] as const;
