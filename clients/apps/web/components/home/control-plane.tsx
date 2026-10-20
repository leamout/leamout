const areas = ["Voice agents", "Calls", "SIP trunks", "AI providers"];

const activities = [
  { name: "Call established", detail: "SIP signaling" },
  { name: "Agent session attached", detail: "Realtime runtime" },
  { name: "Media stream active", detail: "Audio transport" },
];

export function ControlPlane() {
  return (
    <section className="border-b border-border bg-card">
      <div className="mx-auto grid max-w-[1200px] gap-14 px-5 py-24 sm:px-6 lg:grid-cols-[1.1fr_0.9fr] lg:items-center lg:gap-20 lg:px-8 lg:py-28">
        <div className="overflow-hidden rounded-2xl border border-border bg-background">
          <div className="flex items-center justify-between border-b border-border px-5 py-4">
            <p className="font-mono text-[11px] uppercase tracking-widest">
              Control plane
            </p>
            <span className="font-mono text-[10px] text-muted-foreground">
              Illustrative console view
            </span>
          </div>
          <div className="grid grid-cols-[112px_1fr] sm:grid-cols-[150px_1fr]">
            <div className="space-y-1 border-r border-border bg-card p-3">
              {areas.map((area) => (
                <div
                  key={area}
                  className={
                    area === "Calls"
                      ? "rounded-md bg-muted px-3 py-2.5 text-xs font-medium text-foreground"
                      : "px-3 py-2.5 text-xs text-muted-foreground"
                  }
                >
                  {area}
                </div>
              ))}
            </div>
            <div className="min-w-0 p-4 sm:p-6">
              <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                Example call detail
              </p>
              <h3 className="mt-3 text-lg font-semibold">Inbound voice call</h3>
              <div className="mt-4 flex flex-wrap items-center gap-2">
                <span className="size-1.5 rounded-full bg-emerald-400" />
                <span className="text-xs font-medium">Connected</span>
                <span className="text-xs text-muted-foreground">
                  / Agent session
                </span>
              </div>
              <div className="mt-6 border-t border-border pt-5">
                <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                  Activity
                </p>
                <ul className="mt-4 space-y-4">
                  {activities.map((activity) => (
                    <li key={activity.name} className="flex items-start gap-3">
                      <span className="mt-1.5 size-1.5 shrink-0 rounded-full bg-brand-orange" />
                      <div>
                        <p className="text-xs font-medium">{activity.name}</p>
                        <p className="mt-1 text-[11px] text-muted-foreground">
                          {activity.detail}
                        </p>
                      </div>
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          </div>
        </div>
        <div>
          <p className="font-mono text-xs uppercase tracking-[0.16em] text-brand-orange">
            One control plane
          </p>
          <h2 className="mt-5 text-balance text-4xl font-heading font-semibold tracking-tight sm:text-5xl">
            Operate the entire call, not just the prompt.
          </h2>
          <p className="mt-6 leading-8 text-muted-foreground">
            Configure agents, connect SIP trunks, manage provider integrations,
            and inspect what happened during conversations from one place.
          </p>
          <ul className="mt-8 space-y-3 text-sm text-foreground">
            <li className="border-l-2 border-brand-orange pl-4">
              Call activity and diagnostics
            </li>
            <li className="border-l-2 border-brand-orange pl-4">
              Conversation history and recordings
            </li>
            <li className="border-l-2 border-brand-orange pl-4">
              Webhook events and configured tools
            </li>
          </ul>
        </div>
      </div>
    </section>
  );
}
