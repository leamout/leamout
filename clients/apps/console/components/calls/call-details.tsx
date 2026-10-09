"use client";

import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@leamout/ui/components/tabs";
import Link from "next/link";
import { type CallDetailsData, formatDuration } from "@/components/calls/types";

function EmptyState({ children }: { children: React.ReactNode }) {
  return (
    <p className="rounded-lg border border-dashed p-8 text-sm text-muted-foreground">
      {children}
    </p>
  );
}

export function CallDetails({
  callId,
  call,
}: {
  callId: string;
  call?: CallDetailsData;
}) {
  const summary = [
    ["Status", call?.status ?? "Unavailable"],
    ["Direction", call?.direction ?? "—"],
    ["From", call?.from ?? "—"],
    ["To", call?.to ?? "—"],
    ["Started", call?.startedAt ?? "—"],
    ["Ended", call?.endedAt ?? "—"],
    ["Duration", formatDuration(call?.durationSeconds ?? null)],
  ];

  return (
    <div className="space-y-6">
      <Link
        href="/calls"
        className="text-sm text-muted-foreground hover:underline"
      >
        Back to calls
      </Link>
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">Call details</h1>
        <p className="break-all text-sm text-muted-foreground">
          Call ID: {callId}
        </p>
        <p className="text-sm text-muted-foreground">
          UI preview.{" "}
          {call
            ? "Preview data is shown below."
            : "Call data has not been loaded."}
        </p>
      </div>
      <dl className="grid gap-6 rounded-lg border p-6 sm:grid-cols-2 lg:grid-cols-4">
        <div>
          <dt className="text-sm text-muted-foreground">Agent</dt>
          <dd className="mt-1 font-medium">
            {call?.agentId ? (
              <Link
                className="hover:underline"
                href={`/agents/${encodeURIComponent(call.agentId)}`}
              >
                {call.agentName}
              </Link>
            ) : (
              (call?.agentName ?? "—")
            )}
          </dd>
        </div>
        {summary.map(([label, value]) => (
          <div key={label}>
            <dt className="text-sm text-muted-foreground">{label}</dt>
            <dd className="mt-1 break-words font-medium">{value}</dd>
          </div>
        ))}
      </dl>
      {call?.failureReason && (
        <div className="rounded-lg border p-4">
          <h2 className="font-semibold">Failure reason</h2>
          <p className="mt-2 text-sm">{call.failureReason}</p>
        </div>
      )}
      <Tabs defaultValue="conversation" className="gap-6">
        <div className="overflow-x-auto">
          <TabsList aria-label="Call information">
            {[
              "Conversation",
              "Recording",
              "Tools",
              "Events",
              "Diagnostics",
            ].map((label) => (
              <TabsTrigger key={label} value={label.toLowerCase()}>
                {label}
              </TabsTrigger>
            ))}
          </TabsList>
        </div>
        <TabsContent value="conversation">
          {call?.transcript.length ? (
            <ol className="space-y-4" aria-label="Conversation transcript">
              {call.transcript.map((turn) => (
                <li key={turn.id} className="rounded-lg border p-4">
                  <div className="flex justify-between gap-4 text-sm">
                    <span className="font-medium">{turn.speaker}</span>
                    <span className="text-muted-foreground">{turn.time}</span>
                  </div>
                  <p className="mt-2 whitespace-pre-wrap">{turn.text}</p>
                </li>
              ))}
            </ol>
          ) : (
            <EmptyState>No transcript is available.</EmptyState>
          )}
        </TabsContent>
        <TabsContent value="recording">
          {call?.recordingUrl ? (
            <div className="space-y-3">
              <h2 className="font-semibold">Call recording</h2>
              <audio
                controls
                preload="none"
                src={call.recordingUrl}
                className="w-full"
              >
                <track
                  kind="captions"
                  src={call.recordingCaptionsUrl}
                  label="Call transcript"
                  default
                />
                Your browser does not support audio playback.
              </audio>
            </div>
          ) : (
            <EmptyState>No recording is available.</EmptyState>
          )}
        </TabsContent>
        <TabsContent value="tools">
          {call?.tools.length ? (
            <ol className="space-y-4" aria-label="Tool execution history">
              {call.tools.map((tool) => (
                <li key={tool.id} className="space-y-2 rounded-lg border p-4">
                  <h2 className="font-medium">{tool.name}</h2>
                  <p className="text-sm text-muted-foreground">
                    {tool.status} · {tool.time}
                  </p>
                  <p className="whitespace-pre-wrap text-sm">{tool.details}</p>
                </li>
              ))}
            </ol>
          ) : (
            <EmptyState>No tool executions are available.</EmptyState>
          )}
        </TabsContent>
        <TabsContent value="events">
          {call?.events.length ? (
            <ol className="space-y-4" aria-label="Call event timeline">
              {call.events.map((event) => (
                <li key={event.id} className="border-l-2 pl-4">
                  <p className="text-xs text-muted-foreground">{event.time}</p>
                  <h2 className="mt-1 font-medium">{event.title}</h2>
                  <p className="mt-1 text-sm text-muted-foreground">
                    {event.description}
                  </p>
                </li>
              ))}
            </ol>
          ) : (
            <EmptyState>No call events are available.</EmptyState>
          )}
        </TabsContent>
        <TabsContent value="diagnostics">
          <div className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Inspect SIP connectivity, media attachment, AI providers, tool
              execution, and runtime termination.
            </p>
            {call?.diagnostics.length ? (
              <dl className="space-y-4">
                {call.diagnostics.map((item) => (
                  <div key={item.stage} className="rounded-lg border p-4">
                    <dt className="font-medium">{item.stage}</dt>
                    <dd className="mt-2 space-y-1 text-sm">
                      <p>{item.status}</p>
                      <p className="text-muted-foreground">{item.detail}</p>
                    </dd>
                  </div>
                ))}
              </dl>
            ) : (
              <EmptyState>
                Diagnostics have not been loaded. No health or failure status is
                inferred.
              </EmptyState>
            )}
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
