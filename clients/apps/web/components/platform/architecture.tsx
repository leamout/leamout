import Link from "next/link";

const runtime = [
  {
    number: "01",
    title: "SIP signaling",
    description:
      "Receive authorized SIP traffic from configured trunks, resolve inbound routing, and apply call admission rules.",
  },
  {
    number: "02",
    title: "Realtime media",
    description:
      "Connect live audio to an agent session, handle media transport, and return generated speech over the call.",
  },
  {
    number: "03",
    title: "Agent orchestration",
    description:
      "Run configured speech turns, interruptions, provider calls, and permitted tool execution during a session.",
  },
  {
    number: "04",
    title: "Lifecycle and operations",
    description:
      "Persist call state and milestones, expose call details, and deliver configured webhook events.",
  },
];

export function PlatformArchitecture() {
  return (
    <section className="theme-light border-y border-border bg-background text-foreground">
      <div className="mx-auto max-w-6xl px-5 py-20 sm:px-6 lg:px-8 lg:py-28">
        <div className="grid gap-8 lg:grid-cols-[1.05fr_0.95fr] lg:items-end">
          <div>
            <p className="font-mono text-xs uppercase tracking-widest text-foreground">
              Inside the voice path
            </p>
            <h2 className="mt-5 max-w-2xl text-balance font-heading text-4xl font-semibold tracking-tight">
              What runs between SIP and the model.
            </h2>
          </div>
          <p className="max-w-lg leading-8 text-muted-foreground">
            A call is a continuous media session, not a single model request.
            Leamout coordinates telephony, agent execution, and call state
            while you control the configuration.
          </p>
        </div>
        <ol className="mt-12 grid gap-px overflow-hidden rounded-xl border border-border bg-border md:grid-cols-2">
          {runtime.map((layer) => (
            <li key={layer.number} className="bg-card p-7 sm:p-9">
              <span className="font-mono text-xs tabular-nums text-brand-orange">
                {layer.number}
              </span>
              <h3 className="mt-8 text-xl font-semibold">{layer.title}</h3>
              <p className="mt-3 text-sm leading-7 text-muted-foreground">
                {layer.description}
              </p>
            </li>
          ))}
        </ol>
        <div className="mt-4 grid gap-7 rounded-xl border border-border bg-card px-7 py-8 lg:grid-cols-[0.8fr_1.2fr] lg:items-center sm:px-9">
          <div>
            <p className="font-mono text-xs uppercase tracking-widest text-foreground">
              Execution modes
            </p>
            <h3 className="mt-3 text-xl font-semibold">
              Realtime or composable.
            </h3>
          </div>
          <div>
            <p className="text-sm leading-7 text-muted-foreground">
              A realtime agent binds one speech-capable provider. A composable
              agent binds STT, LLM, and TTS separately. Provider readiness is
              checked before calls are accepted for that agent.
            </p>
            <Link
              href="/docs/guides/voice/voice-agents"
              className="mt-5 inline-block text-sm font-medium text-foreground underline decoration-brand-orange underline-offset-4 hover:text-foreground"
            >
              Voice Agent configuration
            </Link>
          </div>
        </div>
      </div>
    </section>
  );
}
