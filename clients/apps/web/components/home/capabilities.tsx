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

const exampleEvents = [
  { step: "01", name: "Speech detected", detail: "Incoming audio" },
  { step: "02", name: "Agent processing", detail: "Model response" },
  { step: "03", name: "Tool invoked", detail: "Application action" },
  { step: "04", name: "Audio streamed", detail: "Return to caller" },
];

export function Capabilities() {
  return (
    <section className="border-b border-border">
      <div className="mx-auto grid max-w-[1200px] gap-14 px-5 py-24 sm:px-6 lg:grid-cols-2 lg:items-center lg:gap-20 lg:px-8 lg:py-28">
        <div>
          <p className="font-mono text-xs uppercase tracking-[0.16em] text-brand-orange">
            Realtime execution
          </p>
          <h2 className="mt-5 text-balance text-4xl font-heading font-semibold tracking-tight sm:text-5xl">
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
          <div className="flex items-center justify-between gap-3">
            <p className="font-mono text-xs uppercase tracking-widest text-foreground">
              Conversation lifecycle
            </p>
            <span className="font-mono text-[10px] uppercase text-muted-foreground">
              Illustrative sequence
            </span>
          </div>
          <ol className="mt-8">
            {exampleEvents.map((event, index) => (
              <li
                key={event.step}
                className="relative flex gap-5 pb-9 last:pb-0"
              >
                {index < exampleEvents.length - 1 && (
                  <span
                    aria-hidden="true"
                    className="absolute left-[15px] top-9 h-[calc(100%-2.25rem)] w-px bg-border"
                  />
                )}
                <span className="relative z-10 flex size-8 shrink-0 items-center justify-center rounded-full border border-border bg-background font-mono text-xs text-brand-orange">
                  {event.step}
                </span>
                <div className="flex-1 border-b border-border pb-6 last:border-0">
                  <p className="font-medium">{event.name}</p>
                  <p className="mt-1 font-mono text-xs text-muted-foreground">
                    {event.detail}
                  </p>
                </div>
              </li>
            ))}
          </ol>
        </div>
      </div>
    </section>
  );
}
