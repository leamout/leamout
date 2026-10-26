import Link from "next/link";

const layers = [
  {
    label: "Your telephony",
    title: "SIP carrier or PBX",
    detail: "Existing numbers, trunks, and routing",
  },
  {
    label: "Leamout managed runtime",
    title: "SIP edge + voice agent",
    detail: "Signaling, realtime media, turns, and tools",
  },
  {
    label: "Your intelligence",
    title: "AI providers",
    detail: "Realtime model or STT, LLM, and TTS",
  },
];

export function HowItWorks() {
  return (
    <section
      id="architecture"
      className="scroll-mt-24 border-b border-border"
      aria-labelledby="home-architecture-title"
    >
      <div className="mx-auto max-w-[1200px] px-5 py-24 sm:px-6 lg:px-8 lg:py-28">
        <div className="grid gap-6 lg:grid-cols-[1fr_0.9fr] lg:items-end">
          <div>
            <p className="font-mono text-xs uppercase tracking-[0.16em] text-brand-orange">
              One managed voice path
            </p>
            <h2
              id="home-architecture-title"
              className="mt-5 max-w-2xl text-balance font-heading text-4xl font-semibold tracking-tight sm:text-5xl"
            >
              Connect the parts. Leamout runs the call.
            </h2>
          </div>
          <p className="max-w-lg leading-8 text-muted-foreground">
            Keep your existing SIP connectivity and AI accounts. Leamout
            operates the signaling, media, and agent session between them.
          </p>
        </div>

        <div className="mt-12 overflow-hidden rounded-xl border border-border bg-card">
          <div className="grid gap-px bg-border lg:grid-cols-3">
            {layers.map((layer, index) => (
              <div
                key={layer.label}
                className={
                  index === 1
                    ? "relative bg-brand-surface px-7 py-9 sm:px-9"
                    : "bg-card px-7 py-9 sm:px-9"
                }
              >
                {index === 1 && (
                  <span
                    aria-hidden="true"
                    className="absolute inset-x-0 top-0 h-1 bg-brand-orange"
                  />
                )}
                <p className="font-mono text-[11px] uppercase tracking-[0.14em] text-brand-orange">
                  {layer.label}
                </p>
                <h3 className="mt-6 font-heading text-2xl font-semibold">
                  {layer.title}
                </h3>
                <p className="mt-3 text-sm leading-7 text-muted-foreground">
                  {layer.detail}
                </p>
              </div>
            ))}
          </div>
          <div className="grid gap-5 border-t border-border bg-background px-7 py-6 md:grid-cols-[0.85fr_1.15fr] md:items-center sm:px-9">
            <div>
              <p className="font-mono text-[11px] uppercase tracking-[0.14em] text-brand-orange">
                Leamout control plane
              </p>
              <h3 className="mt-2 text-lg font-semibold">
                Configure, route, and inspect.
              </h3>
            </div>
            <p className="text-sm leading-7 text-muted-foreground">
              Manage Voice Agents and provider bindings, assign inbound
              numbers, inspect persisted call details, and subscribe to
              lifecycle events. Audio returns to the caller over the SIP
              connection.
            </p>
          </div>
        </div>
        <div className="mt-7 flex flex-wrap items-center justify-between gap-5">
          <p className="max-w-2xl text-xs leading-6 text-muted-foreground">
            Simplified architecture. SIP ingress and AI provider connectivity
            must be configured before an agent can take a call.
          </p>
          <Link
            href="/platform"
            className="shrink-0 text-sm font-medium text-foreground underline decoration-brand-orange underline-offset-4 hover:text-brand-orange"
          >
            Explore the architecture
          </Link>
        </div>
      </div>
    </section>
  );
}
