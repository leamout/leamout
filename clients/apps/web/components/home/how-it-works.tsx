import Link from "next/link";

const responsibilities = [
  {
    number: "01",
    label: "SIGNAL",
    title: "Connect the call",
    detail:
      "Accept SIP signaling and work with your existing carrier, numbers, or PBX.",
  },
  {
    number: "02",
    label: "STREAM",
    title: "Move realtime media",
    detail:
      "Keep the audio path connected while the call and conversation progress.",
  },
  {
    number: "03",
    label: "EXECUTE",
    title: "Run the agent",
    detail:
      "Coordinate your selected realtime model or your STT, LLM, and TTS stack.",
  },
  {
    number: "04",
    label: "OPERATE",
    title: "Track the lifecycle",
    detail:
      "Inspect persisted call milestones, routing identifiers, and recordings where available.",
  },
];

export function HowItWorks() {
  return (
    <section
      id="architecture"
      aria-labelledby="home-architecture-title"
      className="border-b border-border bg-white"
    >
      <div className="mx-auto max-w-[1280px] px-5 py-20 sm:px-6 sm:py-24 lg:px-8 lg:py-28">
        <div className="grid gap-7 lg:grid-cols-[0.9fr_1.1fr] lg:items-end lg:gap-16">
          <div>
            <p className="font-mono text-xs uppercase tracking-[0.16em] text-brand-orange">
              02 / The managed runtime
            </p>
            <h2
              id="home-architecture-title"
              className="mt-5 max-w-2xl text-balance font-heading text-4xl font-semibold leading-tight tracking-[-0.04em] sm:text-5xl"
            >
              A voice call is a system, not just a prompt.
            </h2>
          </div>
          <div className="max-w-2xl lg:justify-self-end">
            <p className="text-base leading-7 text-muted-foreground sm:text-lg sm:leading-8">
              Real calls involve signaling, media, agent execution, and changing
              call state. Leamout provides the runtime between telephony and AI,
              while your team keeps control of the agent experience and provider
              choices.
            </p>
            <Link
              href="/platform"
              className="mt-5 inline-flex items-center gap-2 text-sm font-semibold text-foreground underline decoration-brand-orange decoration-2 underline-offset-4 transition-colors hover:text-brand-orange"
            >
              Explore the platform
              <span aria-hidden="true">↗</span>
            </Link>
          </div>
        </div>

        <ol className="mt-12 grid gap-8 border-y border-border py-8 sm:mt-16 sm:grid-cols-2 sm:gap-x-10 sm:gap-y-10 lg:grid-cols-4 lg:gap-7">
          {responsibilities.map((item) => (
            <li key={item.number} className="min-w-0">
              <div className="flex items-center justify-between border-b border-border pb-4">
                <span className="font-mono text-[10px] uppercase tracking-[0.16em] text-brand-orange">
                  {item.label}
                </span>
                <span className="font-mono text-xs text-muted-foreground">
                  {item.number}
                </span>
              </div>
              <h3 className="mt-5 font-heading text-xl font-semibold tracking-tight">
                {item.title}
              </h3>
              <p className="mt-3 text-sm leading-7 text-muted-foreground">
                {item.detail}
              </p>
            </li>
          ))}
        </ol>
      </div>
    </section>
  );
}
