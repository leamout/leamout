export const recordingStatuses = [
  "recording",
  "uploading",
  "completed",
  "failed",
  "deleted",
] as const;

export type Recording = {
  id: string;
  callId: string;
  status: (typeof recordingStatuses)[number];
  format?: string;
  fileSizeBytes?: number;
  durationSeconds?: number;
  startedAt?: string;
  completedAt?: string;
  createdAt: string;
  updatedAt: string;
};

export type RecordingPlayback = {
  url: string;
  expiresAt: string;
  captionsUrl?: string;
};

export function formatRecordingDate(value?: string) {
  if (!value) return "—";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  return new Intl.DateTimeFormat("en", {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: "UTC",
  }).format(date);
}

export function formatFileSize(bytes?: number) {
  if (bytes === undefined) return "—";
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
