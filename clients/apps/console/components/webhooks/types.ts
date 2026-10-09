export type WebhookEndpoint = {
  id: string;
  url: string;
  enabled: boolean;
  subscribedEvents: string[];
  consecutiveFailures: number;
  disabledReason?: string;
};

export type WebhookDelivery = {
  id: string;
  eventId: string;
  status: string;
  attemptCount: number;
  replayCount: number;
  lastAttemptAt?: string;
  responseStatus?: number;
  responseBody?: string;
  lastError?: string;
};
