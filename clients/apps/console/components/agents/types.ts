export type VoiceAgent = {
  id: string;
  organization_id: string;
  name: string;
  engine: "realtime" | "composable";
  instructions: string;
  voice?: string | null;
  language: string;
  status: "active" | "disabled";
  engine_config: Record<string, unknown>;
  interruption_policy: "allow" | "disabled";
  recording_policy: "none" | "all";
  configuration_revision: number;
  created_at: string;
  updated_at: string;
};

export type VoiceAgentInput = {
  name: string;
  engine: VoiceAgent["engine"];
  instructions: string;
  voice?: string;
  language: string;
  engine_config: Record<string, unknown>;
  interruption_policy: VoiceAgent["interruption_policy"];
  recording_policy: VoiceAgent["recording_policy"];
};
