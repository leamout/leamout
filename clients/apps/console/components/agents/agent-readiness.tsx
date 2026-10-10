export type AgentReadinessReport = {
  ready: boolean;
  configurationRevision: number;
  engine: string;
  bindings: {
    role: string;
    provider: string;
    connectionState: string;
    failureCode?: string;
  }[];
  issues: {
    code: string;
    field: string;
    message: string;
    remediation: string;
  }[];
};

export function AgentReadiness({ report }: { report?: AgentReadinessReport }) {
  return (
    <section
      className="space-y-6 rounded-lg border p-6"
      aria-labelledby="agent-readiness-title"
    >
      <h2 id="agent-readiness-title" className="text-lg font-semibold">
        Readiness
      </h2>
      <p className="font-medium">
        {report
          ? report.ready
            ? "Ready"
            : "Needs attention"
          : "Readiness has not been checked."}
      </p>
      {report && (
        <>
          <p className="text-sm text-muted-foreground">
            Configuration revision {report.configurationRevision} ·{" "}
            {report.engine}
          </p>
          <div className="space-y-3">
            {report.bindings.map((binding) => (
              <div key={binding.role} className="rounded-md border p-4">
                <p className="font-medium">
                  {binding.role} · {binding.provider}
                </p>
                <p className="mt-1 text-sm">{binding.connectionState}</p>
                {binding.failureCode && (
                  <p className="mt-1 text-sm text-muted-foreground">
                    {binding.failureCode}
                  </p>
                )}
              </div>
            ))}
          </div>
          <ul className="space-y-3" aria-label="Readiness issues">
            {report.issues.map((issue) => (
              <li
                key={`${issue.code}-${issue.field}-${issue.message}`}
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
        </>
      )}
      <p className="text-sm text-muted-foreground">
        UI preview. Readiness is not inferred from unsaved form values.
      </p>
    </section>
  );
}
