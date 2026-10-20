const callPath = [
  {
    number: "01",
    label: "Customer connection",
    title: "Your SIP trunk",
    description:
      "Bring existing phone numbers and SIP connectivity from your carrier or PBX.",
  },
  {
    number: "02",
    label: "Telephony and media",
    title: "Leamout SIP edge",
    description:
      "Handle call signaling, media transport, and the connection to an agent session.",
  },
  {
    number: "03",
    label: "Live orchestration",
    title: "Voice agent runtime",
    description:
      "Manage speech turns, interruptions, tool calls, and the ongoing conversation.",
  },
  {
    number: "04",
    label: "Model execution",
    title: "Your AI providers",
    description:
      "Use a realtime speech model or compose recognition, language, and synthesis.",
  },
];

export function PlatformArchitecture() {
  return (
    <section className="marketing-light border-y border-border bg-background text-foreground">
      <div className="mx-auto max-w-6xl px-5 py-20 sm:px-6 lg:px-8">
        <p className="font-mono text-xs uppercase tracking-widest text-foreground">
          How a call moves through Leamout
        </p>
        <h2 className="mt-4 max-w-2xl text-3xl font-semibold tracking-tight sm:text-4xl">
          From a ringing phone to a responsive agent.
        </h2>
        <p className="mt-4 max-w-2xl leading-8 text-muted-foreground">
          The media path and agent session work together while the control plane
          keeps configuration, routing, and activity in one place.
        </p>
        <ol className="mt-12 grid gap-3 md:grid-cols-2 lg:grid-cols-4">
          {callPath.map((step) => (
            <li
              key={step.number}
              className="flex h-full flex-col rounded-xl border border-border border-t-2 border-t-brand-orange bg-card p-6"
            >
              <div className="flex items-center justify-between gap-4">
                <span className="font-mono text-sm text-foreground">
                  {step.number}
                </span>
                <span className="text-xs text-muted-foreground">
                  {step.label}
                </span>
              </div>
              <h3 className="mt-10 text-xl font-semibold">{step.title}</h3>
              <p className="mt-3 text-sm leading-7 text-muted-foreground">
                {step.description}
              </p>
            </li>
          ))}
        </ol>
        <div className="mt-4 grid gap-4 rounded-xl border border-border bg-card p-6 sm:grid-cols-[1fr_2fr] sm:items-center">
          <div>
            <p className="font-mono text-xs uppercase tracking-widest text-foreground">
              Control plane
            </p>
            <h3 className="mt-2 text-lg font-semibold">
              One place to operate every call.
            </h3>
          </div>
          <p className="text-sm leading-7 text-muted-foreground">
            Configure agents and routing, manage credentials and tools, inspect
            call activity, and receive events through webhooks.
          </p>
        </div>
      </div>
    </section>
  );
}
