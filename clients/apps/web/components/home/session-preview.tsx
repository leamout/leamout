import Link from "next/link";

const configuration = [
  { label: "Name", detail: "Name your Voice Agent" },
  { label: "Conversation engine", detail: "Realtime or STT → LLM → TTS" },
  { label: "Instructions", detail: "Define the agent's behavior" },
  { label: "Interruptions", detail: "Allow or disable" },
];

const tabs = ["Configuration", "Readiness", "Providers", "Routing", "Tools"];

export function SessionPreview() {
  return (
    <section
      aria-labelledby="agent-preview-title"
      className="overflow-hidden rounded-2xl border border-border bg-card shadow-[0_24px_90px_rgba(0,0,0,0.2)]"
    >
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border px-5 py-4">
        <h2
          id="agent-preview-title"
          className="font-mono text-[11px] font-medium uppercase tracking-[0.14em] text-foreground"
        >
          Voice Agent configuration
        </h2>
        <span className="font-mono text-[10px] text-muted-foreground">
          Console interface reference
        </span>
      </div>
      <div className="border-b border-border px-4 pt-4 sm:px-5">
        <div className="flex flex-wrap gap-x-5 gap-y-2 text-xs">
          {tabs.map((tab) => (
            <span
              key={tab}
              className={
                tab === "Configuration"
                  ? "border-b-2 border-brand-orange pb-3 font-medium text-foreground"
                  : "pb-3 text-muted-foreground"
              }
            >
              {tab}
            </span>
          ))}
        </div>
      </div>
      <div className="space-y-4 p-5 sm:p-7">
        <p className="text-xs leading-6 text-muted-foreground">
          The fields below reflect the current Console. They are not a customer
          session or a preconfigured agent.
        </p>
        <dl className="grid gap-3 sm:grid-cols-2">
          {configuration.map(({ label, detail }) => (
            <div
              key={label}
              className="min-w-0 rounded-lg border border-border bg-background px-4 py-3"
            >
              <dt className="font-mono text-[10px] uppercase tracking-wider text-muted-foreground">
                {label}
              </dt>
              <dd className="mt-2 text-sm font-medium text-foreground">
                {detail}
              </dd>
            </div>
          ))}
        </dl>
        <div className="rounded-lg border border-border bg-background p-4">
          <div className="flex items-center justify-between gap-4">
            <h3 className="text-sm font-semibold">Agent readiness</h3>
            <span className="rounded-md border border-border px-2 py-1 font-mono text-[10px] text-muted-foreground">
              Configuration check
            </span>
          </div>
          <p className="mt-2 text-xs leading-6 text-muted-foreground">
            The readiness report checks required provider bindings and returns
            issues to resolve. Creating an agent alone does not make it ready to
            receive calls.
          </p>
        </div>
        <div className="border-t border-border pt-4">
          <Link
            href="/docs/guides/voice/voice-agents"
            className="text-sm font-medium text-foreground underline decoration-brand-orange underline-offset-4 hover:text-brand-orange"
          >
            Configure a Voice Agent
          </Link>
        </div>
      </div>
    </section>
  );
}
