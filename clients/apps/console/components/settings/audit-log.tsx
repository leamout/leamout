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
import { useState } from "react";

export type AuditEvent = {
  id: string;
  organizationId: string;
  actorType: string;
  actorId: string;
  action: string;
  targetType: string;
  targetId: string;
  metadata: unknown;
  occurredAt: string;
};

function formatTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  return new Intl.DateTimeFormat("en", {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: "UTC",
  }).format(date);
}

export function AuditLog({ events = [] }: { events?: AuditEvent[] }) {
  const [search, setSearch] = useState("");
  const [actorType, setActorType] = useState("all");
  const [targetType, setTargetType] = useState("all");
  const actorTypes = [
    ...new Set(events.map((event) => event.actorType)),
  ].sort();
  const targetTypes = [
    ...new Set(events.map((event) => event.targetType)),
  ].sort();
  const filtered = events
    .filter(
      (event) =>
        `${event.id} ${event.action} ${event.actorId} ${event.targetId} ${event.targetType}`
          .toLowerCase()
          .includes(search.trim().toLowerCase()) &&
        (actorType === "all" || event.actorType === actorType) &&
        (targetType === "all" || event.targetType === targetType),
    )
    .sort((a, b) => Date.parse(b.occurredAt) - Date.parse(a.occurredAt));

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">Audit log</h1>
        <p className="text-muted-foreground">
          Review administrative activity in your organization, including who
          acted and which resource changed.
        </p>
      </div>
      <div className="flex flex-wrap gap-4">
        <div className="min-w-48 flex-1 space-y-2">
          <label htmlFor="audit-search" className="text-sm font-medium">
            Search events
          </label>
          <Input
            id="audit-search"
            type="search"
            placeholder="Action, actor ID, resource, or event ID"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
          />
        </div>
        <div className="space-y-2">
          <label htmlFor="audit-actor" className="block text-sm font-medium">
            Actor type
          </label>
          <select
            id="audit-actor"
            className="h-9 rounded-md border bg-background px-3 text-sm"
            value={actorType}
            onChange={(event) => setActorType(event.target.value)}
          >
            <option value="all">All actors</option>
            {actorTypes.map((type) => (
              <option key={type} value={type}>
                {type}
              </option>
            ))}
          </select>
        </div>
        <div className="space-y-2">
          <label htmlFor="audit-target" className="block text-sm font-medium">
            Resource type
          </label>
          <select
            id="audit-target"
            className="h-9 rounded-md border bg-background px-3 text-sm"
            value={targetType}
            onChange={(event) => setTargetType(event.target.value)}
          >
            <option value="all">All resources</option>
            {targetTypes.map((type) => (
              <option key={type} value={type}>
                {type}
              </option>
            ))}
          </select>
        </div>
      </div>
      <div className="overflow-hidden rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              {["Occurred (UTC)", "Action", "Actor", "Resource", "Details"].map(
                (label) => (
                  <TableHead key={label}>{label}</TableHead>
                ),
              )}
            </TableRow>
          </TableHeader>
          <TableBody>
            {filtered.map((event) => (
              <TableRow key={event.id}>
                <TableCell className="align-top whitespace-nowrap">
                  {formatTime(event.occurredAt)}
                </TableCell>
                <TableCell className="align-top font-medium">
                  {event.action}
                </TableCell>
                <TableCell className="align-top">
                  <p>{event.actorType}</p>
                  <p className="mt-1 max-w-48 break-all text-xs text-muted-foreground">
                    {event.actorId}
                  </p>
                </TableCell>
                <TableCell className="align-top">
                  <p>{event.targetType}</p>
                  <p className="mt-1 max-w-48 break-all text-xs text-muted-foreground">
                    {event.targetId}
                  </p>
                </TableCell>
                <TableCell className="align-top">
                  <details className="max-w-sm whitespace-normal">
                    <summary className="cursor-pointer font-medium">
                      View event details
                    </summary>
                    <dl className="mt-3 space-y-3 text-xs">
                      <div>
                        <dt className="text-muted-foreground">Event ID</dt>
                        <dd className="mt-1 break-all">{event.id}</dd>
                      </div>
                      <div>
                        <dt className="text-muted-foreground">
                          Organization ID
                        </dt>
                        <dd className="mt-1 break-all">
                          {event.organizationId}
                        </dd>
                      </div>
                      <div>
                        <dt className="text-muted-foreground">Timestamp</dt>
                        <dd className="mt-1 break-all">{event.occurredAt}</dd>
                      </div>
                      <div>
                        <dt className="text-muted-foreground">Metadata</dt>
                        <dd className="mt-1">
                          {event.metadata == null ? (
                            "No metadata available."
                          ) : (
                            <pre className="max-h-64 overflow-auto rounded-md bg-muted p-3 whitespace-pre-wrap break-all">
                              {JSON.stringify(event.metadata, null, 2)}
                            </pre>
                          )}
                        </dd>
                      </div>
                    </dl>
                  </details>
                </TableCell>
              </TableRow>
            ))}
            {filtered.length === 0 && (
              <TableRow>
                <TableCell colSpan={5} className="h-40 text-center">
                  <p className="font-medium">
                    {events.length === 0
                      ? "No audit events to display"
                      : "No matching events"}
                  </p>
                  <p className="mt-2 text-muted-foreground">
                    {events.length === 0
                      ? "Organization activity will appear here."
                      : "Try another search or filter."}
                  </p>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <p className="text-sm text-muted-foreground">
        UI preview. Audit events are not connected. Filters apply to the
        supplied preview events.
      </p>
    </div>
  );
}
