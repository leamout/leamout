import Link from "next/link";

const events = [
  { name: "call.initiated", detail: "A call begins" },
  { name: "call.ringing", detail: "The destination is being alerted" },
  { name: "call.answered", detail: "The call is answered" },
  { name: "call.active", detail: "The active media phase" },
  { name: "call.completed", detail: "The call has ended" },
];

const availableContext = [
  "SIP Call-ID and trunk attribution",
  "Call lifecycle and persisted milestones",
  "Media state and hangup reason",
  "Recording links where available",
];

export function ControlPlane() {
  return (
    <section
      id="console"
      aria-labelledby="home-operations-title"
      className="border-b border-border bg-background"
    >
      <div className="mx-auto grid max-w-[1280px] gap-12 px-5 py-20 sm:px-6 sm:py-24 lg:grid-cols-[0.88fr_1.12fr] lg:items-center lg:gap-16 lg:px-8 lg:py-28">
        <div className="max-w-xl">
          <p className="font-mono text-xs uppercase tracking-[0.16em] text-brand-orange">
            03 / Call operations
          </p>
          <h2
            id="home-operations-title"
            className="mt-5 text-balance font-heading text-4xl font-semibold leading-tight tracking-[-0.04em] sm:text-5xl"
          >
            Know where the call is—not just what the agent said.
          </h2>
          <p className="mt-6 text-base leading-7 text-muted-foreground">
            The control plane gives your team call-level context after setup.
            Inspect persisted lifecycle milestones, SIP and routing identifiers,
            media state, and recordings where available.
          </p>

          <ul className="mt-8 border-t border-border">
            {availableContext.map((item) => (
              <li
                key={item}
                className="flex gap-3 border-b border-border py-3 text-sm leading-6"
              >
                <span
                  aria-hidden="true"
                  className="mt-2 size-1.5 shrink-0 bg-brand-orange"
                />
                {item}
              </li>
            ))}
          </ul>

          <Link
            href="/docs/guides/voice/call-events"
            className="mt-7 inline-flex items-center gap-2 text-sm font-semibold underline decoration-brand-orange decoration-2 underline-offset-4 transition-colors hover:text-brand-orange"
          >
            Read the call event reference
            <span aria-hidden="true">↗</span>
          </Link>
        </div>

        <div className="min-w-0 border border-foreground/15 bg-white">
          <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border px-5 py-4 sm:px-7">
            <div>
              <p className="font-mono text-[10px] uppercase tracking-[0.14em] text-brand-orange">
                Event reference
              </p>
              <h3 className="mt-1 font-heading text-lg font-semibold">
                Call lifecycle
              </h3>
            </div>
            <span className="border border-border px-2 py-1 font-mono text-[9px] uppercase tracking-wider text-muted-foreground">
              Representative names
            </span>
          </div>

          <ol className="px-5 py-2 sm:px-7">
            {events.map((event, index) => (
              <li
                key={event.name}
                className="grid grid-cols-[32px_minmax(0,1fr)] gap-3 border-b border-border py-4 last:border-b-0"
              >
                <span className="flex size-7 items-center justify-center border border-border font-mono text-[10px] text-brand-orange">
                  {String(index + 1).padStart(2, "0")}
                </span>
                <div className="min-w-0">
                  <p className="break-all font-mono text-sm font-medium">
                    {event.name}
                  </p>
                  <p className="mt-1 text-xs leading-5 text-muted-foreground">
                    {event.detail}
                  </p>
                </div>
              </li>
            ))}
          </ol>

          <p className="border-t border-border bg-[#fbfbf9] px-5 py-4 text-[11px] leading-5 text-muted-foreground sm:px-7">
            This is a static reference to selected API event names, not a live
            trace or a promise that every call follows this exact sequence.
            The current call API does not expose SIP message traces, per-turn
            AI latency, transcripts, or tool execution traces.
          </p>
        </div>
      </div>
    </section>
  );
}
