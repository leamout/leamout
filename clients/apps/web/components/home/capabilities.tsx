import Link from "next/link";

const capabilities = [
  {
    title: "Streaming conversations",
    description:
      "Handle incoming audio, agent output, and speech turns as a live session.",
  },
  {
    title: "Tool execution",
    description:
      "Let the agent call approved tools while the session remains under Leamout control.",
  },
  {
    title: "Call control",
    description:
      "Route calls, manage playback, record sessions, and transfer to a person.",
  },
];

const events = [
  { name: "call.initiated", detail: "A call begins" },
  { name: "call.ringing", detail: "The destination rings" },
  { name: "call.answered", detail: "The call is answered" },
  { name: "call.active", detail: "Active media phase" },
  { name: "call.completed", detail: "The call ends successfully" },
];

export function Capabilities() {
  return (
    <section className="border-b border-border">
      <div className="mx-auto grid max-w-[1200px] gap-14 px-5 py-24 sm:px-6 lg:grid-cols-2 lg:items-center lg:gap-20 lg:px-8 lg:py-28">
        <div>
          <p className="font-mono text-xs uppercase tracking-[0.16em] text-brand-orange">
            Realtime execution
          </p>
          <h2 className="mt-5 text-balance font-heading text-4xl font-semibold tracking-tight sm:text-5xl">
            More than a model on a phone call.
          </h2>
          <p className="mt-6 max-w-lg leading-8 text-muted-foreground">
            Voice agents need a runtime that connects conversational
            intelligence with the realities of telephony: streaming media,
            changing call state, interruptions, and external actions.
          </p>
          <div className="mt-9 divide-y divide-border border-t border-border">
            {capabilities.map((capability) => (
              <div key={capability.title} className="py-5">
                <h3 className="font-semibold">{capability.title}</h3>
                <p className="mt-2 text-sm leading-7 text-muted-foreground">
                  {capability.description}
                </p>
              </div>
            ))}
          </div>
        </div>
        <div className="border border-border bg-card p-6 sm:p-8">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <p className="font-mono text-xs uppercase tracking-widest text-foreground">
              Documented call events
            </p>
            <span className="font-mono text-[10px] uppercase text-muted-foreground">
              API event names
            </span>
          </div>
          <ol className="mt-7 space-y-0">
            {events.map((event, index) => (
              <li
                key={event.name}
                className="relative flex gap-4 pb-7 last:pb-0"
              >
                {index < events.length - 1 && (
                  <span
                    aria-hidden="true"
                    className="absolute left-[15px] top-8 h-[calc(100%-2rem)] w-px bg-border"
                  />
                )}
                <span className="relative z-10 flex size-8 shrink-0 items-center justify-center rounded-full border border-border bg-background font-mono text-[10px] text-brand-orange">
                  {String(index + 1).padStart(2, "0")}
                </span>
                <div className="min-w-0 flex-1 border-b border-border pb-5 last:border-0">
                  <p className="break-all font-mono text-sm font-medium">
                    {event.name}
                  </p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {event.detail}
                  </p>
                </div>
              </li>
            ))}
          </ol>
          <p className="mt-5 text-xs leading-6 text-muted-foreground">
            Selected names from the call event catalog, not a captured live
            event sequence. Calls can follow other paths, including failures.
          </p>
          <Link
            href="/docs/guides/voice/call-events"
            className="mt-5 inline-block text-sm font-medium text-foreground underline decoration-brand-orange underline-offset-4 hover:text-brand-orange"
          >
            View call event reference
          </Link>
        </div>
      </div>
    </section>
  );
}
