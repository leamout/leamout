export type WebhookEndpoint = {
  id: string;
  organization_id: string;
  url: string;
  enabled: boolean;
  subscribed_events: string[];
  consecutive_failures: number;
  created_at: string;
  updated_at: string;
  disabled_at?: string | null;
  last_failure_at?: string | null;
  disabled_reason?: string | null;
};

export type WebhookDelivery = {
  id: string;
  event_id: string;
  endpoint_id: string;
  status: string;
  attempt_count: number;
  replay_count: number;
  next_attempt_at: string;
  last_attempt_at?: string | null;
  last_replayed_at?: string | null;
  response_status?: number | null;
  response_body?: string | null;
  last_error?: string | null;
  delivered_at?: string | null;
  created_at: string;
  updated_at: string;
};

export type WebhookSecretResponse = {
  webhook: WebhookEndpoint;
  signing_secret: string;
};
