export type Organization = {
  id: string;
  name: string;
  status: string;
  role: "owner" | "admin" | "member";
};

export type Plan = {
  code: string;
  name: string;
  currency: string;
  amount_minor: number | null;
  billing_interval: string;
};

export type AvailablePlan = Plan & {
  description?: string;
  pricing_type: string;
  status: string;
  limits: Record<string, number>;
};

export type LicenseClaims = {
  license_id: string;
  customer_id: string;
  deployment_id: string;
  valid_from: string;
  expires_at: string;
  grace_until: string;
  max_concurrent_calls: number;
  max_nodes: number;
  features: string[];
};

export type BillingStatus = {
  deployment_mode: "cloud" | "self_hosted";
  billing_configured: boolean;
  checkout_available: boolean;
  plan: Plan | null;
  subscription: {
    provider: string | null;
    status: string;
    current_period_end: string | null;
    cancel_at_period_end: boolean;
    past_due_at: string | null;
    grace_ends_at: string | null;
  } | null;
  limits: {
    max_concurrent_calls: number | null;
    retention_days: number | null;
    monthly_runtime_minutes: number | null;
    max_voice_agents: number | null;
    max_sip_trunks: number | null;
    period_start: string;
    period_end: string;
  };
  usage: {
    used_seconds: number;
    allowance_seconds: number;
    unlimited: boolean;
    period_start: string;
    period_end: string;
  };
  portal_available: boolean;
  license?: {
    state: string;
    license?: LicenseClaims;
  };
};
