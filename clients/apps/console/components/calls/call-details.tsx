"use client";

import { Button } from "@leamout/ui/components/button";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@leamout/ui/components/tabs";
import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { ApiError, api, errorMessage } from "@/lib/api";
import {
  type Call,
  type CallRecording,
  callDurationSeconds,
  formatCallDate,
  formatDuration,
} from "./types";

function Detail({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="min-w-0 space-y-1">
      <dt className="text-sm text-muted-foreground">{label}</dt>
      <dd className="break-all font-medium">{value}</dd>
    </div>
  );
}

export function CallDetails({ callId }: { callId: string }) {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";

  const call = useQuery({
    queryKey: ["calls", organizationId, callId],
    queryFn: ({ signal }) =>
      api<Call>(`/calls/${encodeURIComponent(callId)}`, {
        organizationId,
        signal,
      }),
    enabled: Boolean(organizationId) && Boolean(callId),
  });

  const recordings = useQuery({
    queryKey: ["call-recordings", organizationId, callId],
    queryFn: ({ signal }) =>
      api<{ recordings: CallRecording[] }>(
        "/recordings/?call_id=" +
          encodeURIComponent(callId) +
          "&limit=50&offset=0",
        { organizationId, signal },
      ),
    enabled: Boolean(organizationId) && Boolean(callId) && call.isSuccess,
  });

  if (call.isPending) {
    return <p role="status">Loading call details…</p>;
  }

  if (call.isError) {
    const missing = call.error instanceof ApiError && call.error.status === 404;

    return (
      <div className="space-y-4" role="alert">
        <h1 className="text-2xl font-semibold">
          {missing ? "Call not found" : "Unable to load call"}
        </h1>
        <p className="text-muted-foreground">{errorMessage(call.error)}</p>
        <Button variant="outline" onClick={() => void call.refetch()}>
          Try again
        </Button>
        <Link href="/calls" className="block text-sm underline">
          Back to calls
        </Link>
      </div>
    );
  }

  const item = call.data;
  const timeline = [
    { label: "Created", value: item.created_at },
    { label: "Started", value: item.started_at },
    { label: "Answered", value: item.answered_at },
    { label: "Ended", value: item.ended_at },
    { label: "Last updated", value: item.updated_at },
  ].filter((entry): entry is { label: string; value: string } =>
    Boolean(entry.value),
  );

  return (
    <div className="space-y-8">
      <Link
        href="/calls"
        className="text-sm text-muted-foreground hover:underline"
      >
        Back to calls
      </Link>
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="space-y-2">
          <h1 className="text-3xl font-semibold tracking-tight">
            Call details
          </h1>
          <p className="break-all font-mono text-xs text-muted-foreground">
            {item.id}
          </p>
          <p className="text-sm capitalize text-muted-foreground">
            {item.direction} · {item.state}
          </p>
        </div>
        <Button
          variant="outline"
          disabled={call.isFetching}
          onClick={() => {
            void call.refetch();
            void recordings.refetch();
          }}
        >
          {call.isFetching ? "Refreshing…" : "Refresh"}
        </Button>
      </div>

      <dl className="grid gap-6 rounded-lg border p-6 sm:grid-cols-2 lg:grid-cols-4">
        <Detail label="State" value={item.state} />
        <Detail label="Media state" value={item.media_state} />
        <Detail label="Direction" value={item.direction} />
        <Detail
          label="Talk time"
          value={formatDuration(callDurationSeconds(item))}
        />
        <Detail label="From" value={item.from_uri} />
        <Detail label="To" value={item.to_uri} />
        <Detail label="Created" value={formatCallDate(item.created_at)} />
        <Detail label="Ended" value={formatCallDate(item.ended_at)} />
      </dl>

      {item.hangup_reason && (
        <section className="space-y-2 rounded-lg border p-5">
          <h2 className="font-semibold">
            {item.state === "failed"
              ? "Failure / hangup reason"
              : "Hangup reason"}
          </h2>
          <p className="break-words text-sm">{item.hangup_reason}</p>
        </section>
      )}

      <Tabs defaultValue="timeline" className="gap-6">
        <div className="overflow-x-auto">
          <TabsList aria-label="Call information">
            <TabsTrigger value="timeline">Lifecycle</TabsTrigger>
            <TabsTrigger value="diagnostics">Diagnostics</TabsTrigger>
            <TabsTrigger value="recordings">Recordings</TabsTrigger>
            <TabsTrigger value="conversation">Conversation</TabsTrigger>
          </TabsList>
        </div>

        <TabsContent value="timeline">
          <section className="space-y-4 rounded-lg border p-6">
            <h2 className="text-lg font-semibold">Persisted milestones</h2>
            <p className="text-sm text-muted-foreground">
              These are call record timestamps, not an event-by-event SIP trace.
              Missing timestamps are omitted.
            </p>
            <ol className="space-y-4">
              {timeline.map((event) => (
                <li key={event.label} className="border-l-2 pl-4">
                  <p className="font-medium">{event.label}</p>
                  <p className="text-sm text-muted-foreground">
                    {formatCallDate(event.value)}
                  </p>
                </li>
              ))}
            </ol>
          </section>
        </TabsContent>

        <TabsContent value="diagnostics">
          <section className="space-y-5 rounded-lg border p-6">
            <div className="space-y-2">
              <h2 className="text-lg font-semibold">
                SIP and routing identifiers
              </h2>
              <p className="text-sm text-muted-foreground">
                Persisted call metadata. A missing identifier does not prove
                that the corresponding system failed.
              </p>
            </div>
            <dl className="grid gap-5 sm:grid-cols-2">
              <Detail label="Call ID" value={item.id} />
              <Detail
                label="SIP Call-ID"
                value={item.sip_call_id || "Not recorded"}
              />
              <Detail
                label="Voice Agent"
                value={
                  item.voice_agent_id ? (
                    <Link
                      className="underline underline-offset-4"
                      href={`/agents/${encodeURIComponent(item.voice_agent_id)}`}
                    >
                      {item.voice_agent_id}
                    </Link>
                  ) : (
                    "Not attributed"
                  )
                }
              />
              <Detail
                label="SIP trunk"
                value={
                  item.trunk_id ? (
                    <Link
                      className="underline underline-offset-4"
                      href={`/sip-trunks/${encodeURIComponent(item.trunk_id)}`}
                    >
                      {item.trunk_id}
                    </Link>
                  ) : (
                    "Not attributed"
                  )
                }
              />
              <Detail
                label="Trunk endpoint ID"
                value={item.trunk_endpoint_id || "Not attributed"}
              />
              <Detail label="Media state" value={item.media_state} />
              <Detail
                label="Answered"
                value={formatCallDate(item.answered_at)}
              />
              <Detail label="Ended" value={formatCallDate(item.ended_at)} />
              <Detail
                label="Hangup reason"
                value={item.hangup_reason || "Not recorded"}
              />
            </dl>
            <p className="text-sm text-muted-foreground">
              Media-attachment metrics, SIP message traces, per-turn AI latency,
              and tool execution events are not exposed by this call API.
            </p>
          </section>
        </TabsContent>

        <TabsContent value="recordings">
          <section className="space-y-4 rounded-lg border p-6">
            <h2 className="text-lg font-semibold">Call recordings</h2>
            {recordings.isPending ? (
              <p role="status">Loading recordings…</p>
            ) : recordings.isError ? (
              <div role="alert" className="space-y-3">
                <p>{errorMessage(recordings.error)}</p>
                <Button
                  variant="outline"
                  onClick={() => void recordings.refetch()}
                >
                  Try again
                </Button>
              </div>
            ) : recordings.data.recordings.length ? (
              <ul className="space-y-3">
                {recordings.data.recordings.map((recording) => (
                  <li
                    key={recording.id}
                    className="space-y-2 rounded-md border p-4"
                  >
                    <p className="break-all font-mono text-sm">
                      {recording.id}
                    </p>
                    <p className="text-sm text-muted-foreground">
                      {recording.status} ·{" "}
                      {recording.format ?? "Format unknown"} ·{" "}
                      {formatDuration(recording.duration_seconds ?? null)}
                    </p>
                    <Link
                      className="text-sm underline underline-offset-4"
                      href={`/recordings/${encodeURIComponent(recording.id)}`}
                    >
                      View recording
                    </Link>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="rounded-md border border-dashed p-5 text-sm text-muted-foreground">
                No recordings found for this call.
              </p>
            )}
            {recordings.data?.recordings.length === 50 && (
              <p className="text-xs text-muted-foreground">
                Showing the first 50 recordings for this call.
              </p>
            )}
          </section>
        </TabsContent>

        <TabsContent value="conversation">
          <section className="space-y-3 rounded-lg border p-6">
            <h2 className="text-lg font-semibold">Transcript and tools</h2>
            <p className="text-sm text-muted-foreground">
              Per-call transcripts, turn history, and tool-execution traces are
              not available from the current organization call API. No
              conversation content or tool result is inferred from call state.
            </p>
          </section>
        </TabsContent>
      </Tabs>
    </div>
  );
}
