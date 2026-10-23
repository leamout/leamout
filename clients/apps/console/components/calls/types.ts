export const callStates = [
  "initiating",
  "ringing",
  "answered",
  "active",
  "completed",
  "failed",
  "cancelled",
] as const;

export type CallState = (typeof callStates)[number];

export type Call = {
  id: string;
  organization_id: string;
  voice_agent_id?: string | null;
  trunk_id?: string | null;
  trunk_endpoint_id?: string | null;
  direction: "inbound" | "outbound";
  state: CallState;
  media_state: "active" | "held";
  from_uri: string;
  to_uri: string;
  sip_call_id?: string | null;
  started_at?: string | null;
  answered_at?: string | null;
  ended_at?: string | null;
  hangup_reason?: string | null;
  created_at: string;
  updated_at: string;
};

export type CallRecording = {
  id: string;
  organization_id: string;
  call_id: string;
  status: "recording" | "uploading" | "completed" | "failed" | "deleted";
  file_size_bytes?: number | null;
  format?: string | null;
  duration_seconds?: number | null;
  started_at?: string | null;
  completed_at?: string | null;
  created_at: string;
  updated_at: string;
};

export function formatCallDate(value?: string | null): string {
  if (!value) return "—";

  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? "—"
    : date.toLocaleString(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
      });
}

export function formatDuration(seconds: number | null): string {
  if (seconds === null || !Number.isFinite(seconds)) return "—";

  const total = Math.max(0, Math.floor(seconds));
  const hours = Math.floor(total / 3600);
  const minutes = Math.floor((total % 3600) / 60);
  const remaining = total % 60;

  if (hours) return `${hours}h ${minutes}m ${remaining}s`;
  if (minutes) return `${minutes}m ${remaining}s`;
  return `${remaining}s`;
}

export function callDurationSeconds(call: Call): number | null {
  // A call without an answer or an end has no persisted final talk duration.
  if (!call.answered_at || !call.ended_at) return null;

  const answered = Date.parse(call.answered_at);
  const ended = Date.parse(call.ended_at);
  if (!Number.isFinite(answered) || !Number.isFinite(ended)) return null;

  return Math.max(0, Math.floor((ended - answered) / 1000));
}
