export type CallSummary = {
  id: string;
  agentId?: string;
  agentName: string;
  direction: "inbound" | "outbound";
  status: "queued" | "ringing" | "active" | "completed" | "failed" | "canceled";
  from: string;
  to: string;
  startedAt: string;
  durationSeconds: number | null;
};

export type CallDetailsData = CallSummary & {
  endedAt?: string;
  transcript: { id: string; speaker: string; text: string; time: string }[];
  recordingUrl?: string;
  recordingCaptionsUrl?: string;
  events: { id: string; time: string; title: string; description: string }[];
  tools: {
    id: string;
    name: string;
    status: string;
    time: string;
    details: string;
  }[];
  diagnostics: { stage: string; status: string; detail: string }[];
  failureReason?: string;
};

export function formatDuration(seconds: number | null) {
  if (seconds === null) return "—";
  return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
}
