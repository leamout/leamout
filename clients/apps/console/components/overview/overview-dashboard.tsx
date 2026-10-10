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
import Link from "next/link";
import { type CallSummary, formatDuration } from "@/components/calls/types";

type OverviewData = {
  activeCalls?: number;
  callsToday?: number;
  failedCallsToday?: number;
  enabledAgents?: number;
  recentCalls?: CallSummary[];
  setup?: Partial<Record<"provider" | "trunk" | "agent" | "number", boolean>>;
};

const steps = [
  {
    id: "provider" as const,
    title: "Configure AI providers",
    description: "Review platform defaults or add organization credentials.",
    href: "/ai-providers",
    icon: AiBrain01Icon,
  },
  {
    id: "trunk" as const,
    title: "Connect a SIP trunk",
    description: "Bring your carrier or PBX connection.",
    href: "/sip-trunks",
    icon: LinkSquare02Icon,
  },
  {
    id: "agent" as const,
    title: "Configure an agent",
    description: "Choose instructions, models, voice, and tools.",
    href: "/agents",
    icon: RoboticIcon,
  },
  {
    id: "number" as const,
    title: "Route an existing number",
    description: "Connect a BYOC number and assign its agent.",
    href: "/phone-numbers",
    icon: TelephoneIcon,
  },
];

export function OverviewDashboard({ data = {} }: { data?: OverviewData }) {
  const metrics = [
    {
      title: "Active calls",
      value: data.activeCalls,
      description: "Current organization calls",
    },
    {
      title: "Calls today",
      value: data.callsToday,
      description: "Since 00:00 UTC",
    },
    {
      title: "Failed calls today",
      value: data.failedCallsToday,
      description: "Since 00:00 UTC",
    },
    {
      title: "Enabled agents",
      value: data.enabledAgents,
      description: "Enabled status does not imply readiness",
    },
  ];
  const knownSteps = steps.filter(
    (step) => data.setup?.[step.id] !== undefined,
  ).length;
  const completedSteps = steps.filter(
    (step) => data.setup?.[step.id] === true,
  ).length;

  return (
    <div className="space-y-8">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="space-y-2">
          <h1 className="text-3xl font-semibold tracking-tight">Overview</h1>
          <p className="text-muted-foreground">
            Configure your voice agents and review call activity.
          </p>
        </div>
        <Button render={<Link href="/agents/new" />}>Create agent</Button>
      </div>
      <p className="rounded-lg border bg-muted/30 px-4 py-3 text-sm text-muted-foreground">
        Console preview. Live organization metrics, setup status, and recent
        calls are not connected.
      </p>
      <section
        aria-label="Organization summary"
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
              {metric.value === undefined ? "Not loaded" : metric.description}
            </p>
          </div>
        ))}
      </section>
      <section className="space-y-4" aria-labelledby="setup-title">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <h2 id="setup-title" className="text-xl font-semibold">
            Workspace setup
          </h2>
          <p className="text-sm text-muted-foreground">
            {knownSteps === steps.length
              ? `${completedSteps} of ${steps.length} complete`
              : "Setup progress not loaded"}
          </p>
        </div>
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
                color="currentColor"
                aria-hidden="true"
                className="shrink-0"
              />
              <div className="space-y-2">
                <h3 className="font-medium">{step.title}</h3>
                <p className="text-sm text-muted-foreground">
                  {step.description}
                </p>
                <p className="text-xs font-medium">
                  {data.setup?.[step.id] === undefined
                    ? "Status not loaded"
                    : data.setup[step.id]
                      ? "Complete"
                      : "Setup required"}
                </p>
              </div>
            </Link>
          ))}
        </div>
      </section>
      <section className="space-y-4" aria-labelledby="recent-calls-title">
        <div className="flex items-center justify-between gap-4">
          <h2 id="recent-calls-title" className="text-xl font-semibold">
            Recent calls
          </h2>
          <Link href="/calls" className="text-sm underline underline-offset-4">
            View all calls
          </Link>
        </div>
        <div className="overflow-hidden rounded-xl border">
          <Table>
            <TableHeader>
              <TableRow>
                {[
                  "Call",
                  "Agent",
                  "Direction",
                  "Status",
                  "Started",
                  "Duration",
                ].map((label) => (
                  <TableHead key={label}>{label}</TableHead>
                ))}
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.recentCalls?.map((call) => (
                <TableRow key={call.id}>
                  <TableCell>
                    <Link
                      href={`/calls/${encodeURIComponent(call.id)}`}
                      className="font-medium hover:underline"
                    >
                      {call.id}
                    </Link>
                  </TableCell>
                  <TableCell>{call.agentName}</TableCell>
                  <TableCell>{call.direction}</TableCell>
                  <TableCell>{call.status}</TableCell>
                  <TableCell>{call.startedAt}</TableCell>
                  <TableCell>{formatDuration(call.durationSeconds)}</TableCell>
                </TableRow>
              ))}
              {!data.recentCalls?.length && (
                <TableRow>
                  <TableCell colSpan={6} className="h-40 text-center">
                    <HugeiconsIcon
                      icon={Call02Icon}
                      size={24}
                      strokeWidth={1.5}
                      aria-hidden="true"
                      className="mx-auto mb-3 text-muted-foreground"
                    />
                    <p className="font-medium">
                      {data.recentCalls
                        ? "No recent calls"
                        : "Recent calls have not been loaded"}
                    </p>
                    <p className="mt-2 text-sm text-muted-foreground">
                      Review conversations, recordings, and diagnostics from
                      call details.
                    </p>
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </section>
    </div>
  );
}
