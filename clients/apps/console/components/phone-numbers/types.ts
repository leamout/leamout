export type ResourceOption = { id: string; name: string };

export type PhoneNumber = {
  id: string;
  organization_id: string;
  number: string;
  country_code: string;
  trunk_id?: string | null;
  voice_enabled: boolean;
  status: string;
  created_at: string;
  updated_at: string;
};

export type AgentNumberBinding = {
  id: string;
  organization_id: string;
  voice_agent_id: string;
  phone_number_id: string;
  created_at: string;
};
