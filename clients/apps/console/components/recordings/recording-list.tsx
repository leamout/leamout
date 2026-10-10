"use client";

import { Input } from "@leamout/ui/components/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@leamout/ui/components/table";
import Link from "next/link";
import { useState } from "react";
import { formatDuration } from "@/components/calls/types";
import {
  formatFileSize,
  formatRecordingDate,
  type Recording,
  recordingStatuses,
} from "@/components/recordings/types";

export function RecordingList({
  recordings = [],
}: {
  recordings?: Recording[];
}) {
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState("all");
  const filtered = recordings.filter(
    (recording) =>
      `${recording.id} ${recording.callId}`
        .toLowerCase()
        .includes(search.trim().toLowerCase()) &&
      (status === "all" || recording.status === status),
  );

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">Recordings</h1>
        <p className="text-muted-foreground">
          Review recorded call audio and open the associated call.
        </p>
      </div>
      <div className="flex flex-wrap gap-4">
        <div className="min-w-48 flex-1 space-y-2">
          <label htmlFor="recording-search" className="text-sm font-medium">
            Search recordings
          </label>
          <Input
            id="recording-search"
            type="search"
            placeholder="Recording ID or call ID"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
          />
        </div>
        <div className="space-y-2">
          <label
            htmlFor="recording-status"
            className="block text-sm font-medium"
          >
            Status
          </label>
          <select
            id="recording-status"
            className="h-9 rounded-md border bg-background px-3 text-sm"
            value={status}
            onChange={(event) => setStatus(event.target.value)}
          >
            <option value="all">All statuses</option>
            {recordingStatuses.map((value) => (
              <option key={value} value={value}>
                {value}
              </option>
            ))}
          </select>
        </div>
      </div>
      <div className="overflow-hidden rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              {[
                "Recording",
                "Call",
                "Status",
                "Created (UTC)",
                "Duration",
                "Format",
                "Size",
              ].map((label) => (
                <TableHead key={label}>{label}</TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {filtered.map((recording) => (
              <TableRow key={recording.id}>
                <TableCell>
                  <Link
                    className="font-medium hover:underline"
                    href={`/recordings/${encodeURIComponent(recording.id)}`}
                  >
                    {recording.id}
                  </Link>
                </TableCell>
                <TableCell>
                  <Link
                    className="hover:underline"
                    href={`/calls/${encodeURIComponent(recording.callId)}`}
                  >
                    {recording.callId}
                  </Link>
                </TableCell>
                <TableCell className="capitalize">{recording.status}</TableCell>
                <TableCell>
                  {formatRecordingDate(recording.createdAt)}
                </TableCell>
                <TableCell>
                  {formatDuration(recording.durationSeconds ?? null)}
                </TableCell>
                <TableCell>{recording.format?.toUpperCase() ?? "—"}</TableCell>
                <TableCell>{formatFileSize(recording.fileSizeBytes)}</TableCell>
              </TableRow>
            ))}
            {filtered.length === 0 && (
              <TableRow>
                <TableCell colSpan={7} className="h-40 text-center">
                  <p className="font-medium">
                    {recordings.length === 0
                      ? "No recordings to display"
                      : "No matching recordings"}
                  </p>
                  <p className="mt-2 text-muted-foreground">
                    {recordings.length === 0
                      ? "Recordings from your calls will appear here."
                      : "Try another search or status filter."}
                  </p>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <p className="text-sm text-muted-foreground">
        UI preview. Recording data is not connected.
      </p>
    </div>
  );
}
