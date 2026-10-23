"use client";

import {
  AiBrain01Icon,
  Call02Icon,
  LinkSquare02Icon,
  RoboticIcon,
  TelephoneIcon,
} from "@hugeicons/core-free-icons";
import { HugeiconsIcon } from "@hugeicons/react";
import { Button } from "@leamout/ui/components/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@leamout/ui/components/table";
import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { useOrganization } from "@/components/onboarding/organization-provider";
import type { AIIntegration } from "@/components/ai-providers/types";
import type { VoiceAgent } from "@/components/agents/types";
import {
  callDurationSeconds,
  formatCallDate,
  formatDuration,
  type Call,
} from "@/components/calls/types";
import type { PhoneNumber } from "@/components/phone-numbers/types";
import type { SipTrunk } from "@/components/sip-trunks/types";
import { api, errorMessage } from "@/lib/api";

const steps = [
  {
    id: "provider" as const,
    title: "Verify AI credentials",
    description: "Connect an organization-owned AI integration.",
    href: "/ai-providers",
    icon: AiBrain01Icon,
  },
  {
    id: "trunk" as const,
    title: "Connect a SIP trunk",
    description: "Add your carrier or PBX connection.",
    href: "/sip-trunks",
    icon: LinkSquare02Icon,
  },
  {
    id: "agent" as const,
    title: "Configure a Voice Agent",
    description: "Create an active Agent and verify its readiness.",
    href: "/agents",
    icon: RoboticIcon,
  },
  {
    id: "number" as const,
    title: "Connect an existing phone number",
    description: "Register a BYOC number and configure its routing.",
    href: "/phone-numbers",
    icon: TelephoneIcon,
  },
];

export function OverviewDashboard() {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";

  const agents = useQuery({
    queryKey: ["voice-agents", organizationId],
    queryFn: ({ signal }) =>
      api<{ voice_agents: VoiceAgent[] }>("/voice-agents/", {
        organizationId,
        signal,
      }),
    enabled: Boolean(organizationId),
  });

  const trunks = useQuery({
    queryKey: ["trunks", organizationId],
    queryFn: ({ signal }) =>
      api<{ trunks: SipTrunk[] }>("/trunks/", {
        organizationId,
        signal,
      }),
    enabled: Boolean(organizationId),
  });

  const numbers = useQuery({
    queryKey: ["numbers", organizationId],
    queryFn: ({ signal }) =>
      api<{ numbers: PhoneNumber[] }>("/numbers/", {
        organizationId,
        signal,
      }),
    enabled: Boolean(organizationId),
  });

  const integrations = useQuery({
    queryKey: ["ai-integrations", organizationId],
    queryFn: ({ signal }) =>
      api<{ ai_integrations: AIIntegration[] }>("/ai-integrations/", {
        organizationId,
        signal,
      }),
    enabled: Boolean(organizationId),
  });

  const calls = useQuery({
    queryKey: ["overview-calls", organizationId],
    queryFn: ({ signal }) =>
      api<{ calls: Call[] }>("/calls/?limit=10&offset=0", {
        organizationId,
        signal,
      }),
    enabled: Boolean(organizationId),
  });

  const readyIntegrations =
    integrations.data?.ai_integrations.filter(
      (entry) => entry.connection_state === "ready",
    ) ?? [];
  const activeTrunks =
    trunks.data?.trunks.filter((trunk) => trunk.status === "active") ?? [];
  const activeAgents =
    agents.data?.voice_agents.filter((agent) => agent.status === "active") ?? [];
  const registeredNumbers = numbers.data?.numbers ?? [];

  const metrics = [
    {
      title: "Active Voice Agents",
      value: agents.data ? activeAgents.length : undefined,
      description: "Configured active Agents, not readiness",
    },
    {
      title: "Active SIP trunks",
      value: trunks.data ? activeTrunks.length : undefined,
      description: "Configured trunks, not carrier health",
    },
    {
      title: "Registered numbers",
      value: numbers.data ? registeredNumbers.length : undefined,
      description: "Existing BYOC numbers",
    },
    {
      title: "Verified AI credentials",
      value: integrations.data ? readyIntegrations.length : undefined,
      description: "Organization integrations marked ready",
    },
  ];

  const setup = {
    provider: integrations.data ? readyIntegrations.length > 0 : null,
    trunk: trunks.data ? activeTrunks.length > 0 : null,
    agent: agents.data ? activeAgents.length > 0 : null,
    number: numbers.data ? registeredNumbers.length > 0 : null,
  };

  const completed = Object.values(setup).filter((value) => value === true).length;
  const known = Object.values(setup).filter((value) => value !== null).length;

  const errors = [
    agents.error,
    trunks.error,
    numbers.error,
    integrations.error,
  ].filter((error) => error !== null);

  return (
    <div className="space-y-8">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="space-y-2">
          <h1 className="text-3xl font-semibold tracking-tight">Overview</h1>
          <p className="text-muted-foreground">
            {selected
              ? "Organization: " + selected.name
              : "Select an organization to review your workspace."}
          </p>
        </div>
        <Button render={<Link href="/agents/new" />}>Create Agent</Button>
      </div>

      {errors.length > 0 && (
        <div role="alert" className="space-y-3 rounded-lg border p-4">
          <p>
            Some workspace data could not be loaded:{" "}
            {errorMessage(errors[0])}
          </p>
          <Button
            variant="outline"
            onClick={() => {
              void agents.refetch();
              void trunks.refetch();
              void numbers.refetch();
              void integrations.refetch();
            }}
          >
            Retry workspace data
          </Button>
        </div>
      )}

      <section
        aria-label="Organization resources"
        className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4"
      >
        {metrics.map((metric) => (
          <div
            key={metric.title}
            className="space-y-3 rounded-xl border bg-card p-6"
          >
            <h2 className="text-sm font-medium">{metric.title}</h2>
            <p className="text-3xl font-semibold tabular-nums">
              {metric.value === undefined
                ? "—"
                : metric.value.toLocaleString("en-US")}
            </p>
            <p className="text-xs text-muted-foreground">
              {metric.description}
            </p>
          </div>
        ))}
      </section>

      <section className="space-y-4" aria-labelledby="setup-title">
        <div className="flex flex-wrap justify-between gap-2">
          <h2 id="setup-title" className="text-xl font-semibold">
            Workspace setup
          </h2>
          <p className="text-sm text-muted-foreground">
            {known === steps.length
              ? `${completed} of ${steps.length} configured`
              : "Loading setup status…"}
          </p>
        </div>
        <p className="text-sm text-muted-foreground">
          These steps indicate configuration, not end-to-end call readiness.
        </p>
        <div className="grid gap-4 md:grid-cols-2">
          {steps.map((step) => (
            <Link
              key={step.id}
              href={step.href}
              className="flex gap-4 rounded-xl border bg-card p-6 transition-colors hover:bg-accent focus-visible:outline-2 focus-visible:outline-ring"
            >
              <HugeiconsIcon
                icon={step.icon}
                size={24}
                strokeWidth={1.5}
                aria-hidden="true"
                className="shrink-0"
              />
              <div className="space-y-2">
                <h3 className="font-medium">{step.title}</h3>
                <p className="text-sm text-muted-foreground">
                  {step.description}
                </p>
                <p className="text-xs font-medium">
                  {setup[step.id] === null
                    ? "Not loaded"
                    : setup[step.id]
                      ? "Configured"
                      : "Setup required"}
                </p>
              </div>
            </Link>
          ))}
        </div>
      </section>

      <section className="space-y-4" aria-labelledby="recent-calls-title">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 id="recent-calls-title" className="text-xl font-semibold">
            Recent calls
          </h2>
          <div className="flex gap-3">
            <Button
              variant="outline"
              disabled={calls.isFetching || !organizationId}
              onClick={() => void calls.refetch()}
            >
              {calls.isFetching ? "Refreshing…" : "Refresh"}
            </Button>
            <Link href="/calls" className="text-sm underline underline-offset-4">
              View all calls
            </Link>
          </div>
        </div>
        <p className="text-sm text-muted-foreground">
          The latest 10 call records, not an aggregate usage or billing report.
        </p>
        {calls.isPending ? (
          <p role="status">Loading recent calls…</p>
        ) : calls.isError ? (
          <p role="alert">{errorMessage(calls.error)}</p>
        ) : (
          <div className="overflow-x-auto rounded-xl border">
            <Table>
              <TableHeader>
                <TableRow>
                  {["Call", "Agent", "Direction", "State", "Started", "Talk time"].map(
                    (label) => (
                      <TableHead key={label}>{label}</TableHead>
                    ),
                  )}
                </TableRow>
              </TableHeader>
              <TableBody>
                {calls.data.calls.map((call) => (
                  <TableRow key={call.id}>
                    <TableCell>
                      <Link
                        href={"/calls/" + encodeURIComponent(call.id)}
                        className="font-mono text-xs hover:underline"
                      >
                        {call.id}
                      </Link>
                    </TableCell>
                    <TableCell>
                      {call.voice_agent_id
                        ? activeAgents.find(
                            (agent) => agent.id === call.voice_agent_id,
                          )?.name ?? call.voice_agent_id
                        : "—"}
                    </TableCell>
                    <TableCell>{call.direction}</TableCell>
                    <TableCell>{call.state}</TableCell>
                    <TableCell>
                      {formatCallDate(call.started_at ?? call.created_at)}
                    </TableCell>
                    <TableCell>
                      {formatDuration(callDurationSeconds(call))}
                    </TableCell>
                  </TableRow>
                ))}
                {calls.data.calls.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={6} className="py-12 text-center">
                      <HugeiconsIcon
                        icon={Call02Icon}
                        size={24}
                        aria-hidden="true"
                        className="mx-auto mb-3 text-muted-foreground"
                      />
                      No calls recorded for this organization.
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        )}
      </section>
    </div>
  );
}
