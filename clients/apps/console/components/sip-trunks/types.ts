export type SipTrunk = {
  id: string;
  organization_id: string;
  name: string;
  direction: "inbound" | "outbound" | "bidirectional";
  status: "active" | "disabled";
  outbound_auth_method: "none" | "digest";
  outbound_username?: string | null;
  outbound_realm?: string | null;
  has_outbound_credentials: boolean;
  inbound_enabled: boolean;
  inbound_auth_method: "ip" | "digest";
  inbound_username?: string | null;
  inbound_realm?: string | null;
  has_inbound_credentials: boolean;
  max_cps: number;
  max_concurrent_calls: number;
  codecs: string[];
  supports_video: boolean;
  supports_fax: boolean;
  created_at: string;
  updated_at: string;
};

export type TrunkEndpoint = {
  id: string;
  trunk_id: string;
  host: string;
  port: number;
  transport: "udp" | "tcp" | "tls";
  direction: string;
  priority: number;
  weight: number;
  enabled: boolean;
  health_status: string;
  consecutive_failures: number;
  last_checked_at?: string | null;
  last_response_code?: number | null;
  last_latency_ms?: number | null;
  last_error?: string | null;
};

export type TrunkSourceIP = {
  id: string;
  trunk_id: string;
  cidr: string;
  created_at: string;
};

export type TrunkValidation = {
  valid: boolean;
  errors: string[];
  warnings: string[];
};
