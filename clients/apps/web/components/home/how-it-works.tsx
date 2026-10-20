import Link from "next/link";

const callPath = [
  {
    number: "01",
    title: "Your SIP network",
    description: "Existing carrier, PBX, and phone numbers.",
  },
  {
    number: "02",
    title: "Leamout SIP edge",
    description: "Call signaling and realtime media.",
  },
  {
    number: "03",
    title: "Voice agent runtime",
    description: "Session orchestration, turns, and tools.",
  },
  {
    number: "04",
    title: "Your AI providers",
    description: "Realtime models or STT, LLM, and TTS.",
  },
];

export function HowItWorks() {
  return (
    <section className="border-b border-border bg-[#26211F]">
      <div className="mx-auto max-w-[1200px] px-5 py-24 sm:px-6 lg:px-8 lg:py-28">
        <div className="grid gap-8 lg:grid-cols-[1.1fr_0.9fr] lg:items-end">
          <div>
            <p className="font-mono text-xs uppercase tracking-[0.16em] text-[#FF8E6D]">
              A call through Leamout
            </p>
            <h2 className="mt-5 max-w-2xl text-balance text-4xl font-semibold tracking-tight sm:text-5xl">
              From the phone network to a live conversation.
            </h2>
          </div>
          <p className="max-w-lg leading-8 text-muted-foreground">
            You bring SIP connectivity and AI credentials. Leamout operates the
            signaling, media, session orchestration, and controls that connect
            them.
          </p>
        </div>
        <ol className="mt-14 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          {callPath.map((step) => (
            <li
              key={step.number}
              className="relative min-h-56 border border-border bg-background p-6"
            >
              <span
                aria-hidden="true"
                className="absolute inset-x-0 top-0 h-0.5 bg-brand-orange"
              />
              <span className="font-mono text-xs text-[#FF8E6D]">
                {step.number}
              </span>
              <h3 className="mt-12 text-xl font-semibold">{step.title}</h3>
              <p className="mt-3 text-sm leading-7 text-muted-foreground">
                {step.description}
              </p>
            </li>
          ))}
        </ol>
        <div className="mt-3 flex flex-col gap-4 border border-border bg-card px-6 py-5 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <p className="font-mono text-[11px] uppercase tracking-widest text-[#FF8E6D]">
              Leamout control plane
            </p>
            <p className="mt-2 text-sm text-muted-foreground">
              Configure agents, manage routing, inspect calls, and receive events.
            </p>
          </div>
          <Link
            href="/platform"
            className="shrink-0 text-sm font-medium text-foreground underline decoration-brand-orange underline-offset-8 hover:text-[#FF8E6D]"
          >
            Explore the platform
          </Link>
        </div>
      </div>
    </section>
  );
}
