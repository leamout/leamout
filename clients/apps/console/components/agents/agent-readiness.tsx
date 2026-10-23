"use client";

import { Button } from "@leamout/ui/components/button";
import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import type { AgentReadinessReport } from "@/components/ai-providers/types";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { api, errorMessage } from "@/lib/api";

export function AgentReadiness({ agentId }: { agentId: string }) {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";

  const readiness = useQuery({
    queryKey: ["agent-readiness", organizationId, agentId],
    queryFn: ({ signal }) =>
      api<AgentReadinessReport>(
        `/voice-agents/${encodeURIComponent(agentId)}/readiness`,
        { organizationId, signal },
      ),
    enabled: Boolean(organizationId) && Boolean(agentId),
  });

  return (
    <section
      className="space-y-6 rounded-lg border p-6"
      aria-labelledby="agent-readiness-title"
    >
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 id="agent-readiness-title" className="text-lg font-semibold">
          Readiness
        </h2>
        <Button
          variant="outline"
          disabled={readiness.isFetching || !organizationId}
          onClick={() => void readiness.refetch()}
        >
          {readiness.isFetching ? "Checking…" : "Check readiness"}
        </Button>
      </div>
      {readiness.isPending ? (
        <p role="status">Loading Agent readiness…</p>
      ) : readiness.isError ? (
        <div role="alert" className="space-y-3">
          <p>{errorMessage(readiness.error)}</p>
          <Button variant="outline" onClick={() => void readiness.refetch()}>
            Try again
          </Button>
        </div>
      ) : (
        <>
          <p className="font-medium">
            {readiness.data.ready ? "Ready" : "Needs attention"}
          </p>
          <p className="text-sm text-muted-foreground">
            Configuration revision {readiness.data.configuration_revision} ·{" "}
            {readiness.data.engine}
          </p>
          <div className="space-y-3">
            {readiness.data.bindings.map((binding) => (
              <div key={binding.role} className="rounded-md border p-4">
                <p className="font-medium">
                  {binding.role.toUpperCase()} · {binding.provider}
                </p>
                <p className="mt-1 text-sm">{binding.connection_state}</p>
                {binding.failure_code && (
                  <p className="mt-1 text-sm text-muted-foreground">
                    {binding.failure_code}
                  </p>
                )}
              </div>
            ))}
          </div>
          {readiness.data.issues.length ? (
            <ul className="space-y-3" aria-label="Readiness issues">
              {readiness.data.issues.map((issue) => (
                <li
                  key={issue.code + issue.field + issue.message}
                  className="rounded-md border p-4"
                >
                  <p className="font-medium">{issue.message}</p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {issue.code} · {issue.field}
                  </p>
                  <p className="mt-2 text-sm">{issue.remediation}</p>
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-sm text-muted-foreground">
              No readiness issues reported.
            </p>
          )}
          <Link href="/ai-providers" className="text-sm underline">
            Manage AI integrations
          </Link>
        </>
      )}
    </section>
  );
}
