"use client";

import { Button } from "@leamout/ui/components/button";
import Link from "next/link";
import { useState } from "react";
import { formatDuration } from "@/components/calls/types";
import {
  formatFileSize,
  formatRecordingDate,
  type Recording,
  type RecordingPlayback,
} from "@/components/recordings/types";

export function RecordingDetails({
  recordingId,
  recording,
  playback,
}: {
  recordingId: string;
  recording?: Recording;
  playback?: RecordingPlayback;
}) {
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [notice, setNotice] = useState("");
  const [playbackFailed, setPlaybackFailed] = useState(false);
  const [playbackExpired, setPlaybackExpired] = useState(false);
  const expiresAt = playback ? Date.parse(playback.expiresAt) : Number.NaN;
  const hasPlayback = Boolean(
    recording?.status === "completed" &&
      playback?.url &&
      Number.isFinite(expiresAt) &&
      expiresAt > Date.now() &&
      !playbackExpired,
  );
  const metadata = [
    ["Status", recording?.status ?? "Unavailable"],
    ["Duration", formatDuration(recording?.durationSeconds ?? null)],
    ["Format", recording?.format?.toUpperCase() ?? "—"],
    ["File size", formatFileSize(recording?.fileSizeBytes)],
    ["Created (UTC)", formatRecordingDate(recording?.createdAt)],
    ["Started (UTC)", formatRecordingDate(recording?.startedAt)],
    ["Completed (UTC)", formatRecordingDate(recording?.completedAt)],
    ["Updated (UTC)", formatRecordingDate(recording?.updatedAt)],
  ];

  return (
    <div className="space-y-6">
      <Link
        href="/recordings"
        className="text-sm text-muted-foreground hover:underline"
      >
        Back to recordings
      </Link>
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">
          Recording details
        </h1>
        <p className="break-all text-sm text-muted-foreground">
          Recording ID: {recordingId}
        </p>
        <p className="text-sm text-muted-foreground">
          UI preview.{" "}
          {recording
            ? "Preview data is shown below."
            : "Recording data has not been loaded."}
        </p>
      </div>
      <dl className="grid gap-6 rounded-lg border p-6 sm:grid-cols-2 lg:grid-cols-3">
        <div>
          <dt className="text-sm text-muted-foreground">Call</dt>
          <dd className="mt-1 break-all font-medium">
            {recording ? (
              <Link
                className="hover:underline"
                href={`/calls/${encodeURIComponent(recording.callId)}`}
              >
                {recording.callId}
              </Link>
            ) : (
              "—"
            )}
          </dd>
        </div>
        {metadata.map(([label, value]) => (
          <div key={label}>
            <dt className="text-sm text-muted-foreground">{label}</dt>
            <dd className="mt-1 break-words font-medium">{value}</dd>
          </div>
        ))}
      </dl>
      <section
        aria-labelledby="recording-playback-title"
        className="space-y-4 rounded-lg border p-6"
      >
        <h2 id="recording-playback-title" className="font-semibold">
          Playback
        </h2>
        {hasPlayback && !playbackFailed ? (
          <div className="space-y-3">
            <audio
              controls
              preload="none"
              src={playback?.url}
              className="w-full"
              aria-label="Call recording"
              onError={() => setPlaybackFailed(true)}
              onPlay={(event) => {
                if (Date.now() >= expiresAt) {
                  event.currentTarget.pause();
                  setPlaybackExpired(true);
                }
              }}
            >
              <track
                kind="captions"
                src={playback?.captionsUrl}
                label="Call transcript"
                default
              />
              Your browser does not support audio playback.
            </audio>
            <p className="text-sm text-muted-foreground">
              Playback link expires {formatRecordingDate(playback?.expiresAt)}{" "}
              UTC.
            </p>
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">
            {playbackFailed
              ? "Audio could not be loaded. A fresh playback link may be needed."
              : recording?.status === "recording" ||
                  recording?.status === "uploading"
                ? "Audio will be available when recording and upload are complete."
                : recording?.status === "failed"
                  ? "This recording failed. Audio is unavailable."
                  : recording?.status === "deleted"
                    ? "This recording has been deleted."
                    : playback
                      ? "The playback link is expired or unavailable."
                      : "No playback link is available."}
          </p>
        )}
      </section>
      <section
        aria-labelledby="recording-delete-title"
        className="space-y-4 rounded-lg border p-6"
      >
        <h2 id="recording-delete-title" className="font-semibold">
          Delete recording
        </h2>
        <p className="text-sm text-muted-foreground">
          Deleting a recording removes its audio. The associated call remains in
          call history.
        </p>
        {confirmDelete ? (
          <div className="space-y-3">
            <p className="text-sm font-medium">
              Confirm deletion of recording {recordingId}?
            </p>
            <div className="flex flex-wrap gap-3">
              <Button
                variant="destructive"
                onClick={() => {
                  setConfirmDelete(false);
                  setNotice(
                    "Deletion preview only. No recording has been deleted.",
                  );
                }}
              >
                Confirm deletion preview
              </Button>
              <Button variant="outline" onClick={() => setConfirmDelete(false)}>
                Cancel
              </Button>
            </div>
          </div>
        ) : (
          <Button
            variant="destructive"
            disabled={!recording || recording.status === "deleted"}
            onClick={() => {
              setNotice("");
              setConfirmDelete(true);
            }}
          >
            Preview deletion
          </Button>
        )}
        <p role="status" className="text-sm text-muted-foreground">
          {notice}
        </p>
      </section>
    </div>
  );
}
