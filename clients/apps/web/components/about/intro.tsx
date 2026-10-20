const callLayers = [
  {
    number: "01",
    label: "Connect",
    title: "Your SIP carrier",
    detail: "Bring your trunks, numbers, and call routes.",
  },
  {
    number: "02",
    label: "Operate",
    title: "Leamout",
    detail: "Run realtime media, agents, tools, and call sessions.",
  },
  {
    number: "03",
    label: "Reason and respond",
    title: "Your AI providers",
    detail: "Connect speech and language models.",
  },
];

export function AboutIntro() {
  return (
    <section className="border-b border-border py-20 sm:py-24 lg:py-28">
      <div className="grid gap-14 lg:grid-cols-[1.05fr_0.95fr] lg:items-center lg:gap-16">
        <div>
          <p className="mb-7 flex items-center gap-3 font-mono text-xs uppercase tracking-[0.16em] text-foreground">
            <span
              aria-hidden="true"
              className="size-2 rounded-full bg-brand-orange"
            />
            About Leamout
          </p>
          <h1 className="max-w-2xl text-balance font-heading text-5xl font-semibold leading-[1.06] tracking-tight sm:text-6xl">
            Great voice agents need more than a model.
          </h1>
          <p className="mt-8 max-w-xl text-lg leading-8 text-muted-foreground">
            A phone call is a live system. Signaling, audio, interruptions,
            model responses, and tools have to work together while someone is
            on the line.
          </p>
          <p className="mt-5 max-w-xl text-lg leading-8 text-muted-foreground">
            We are building Leamout to make that infrastructure accessible
            through one managed runtime and control plane.
          </p>
        </div>

        <div className="overflow-hidden rounded-xl border border-border bg-brand-surface">
          <div className="flex items-center justify-between gap-4 border-b border-border px-6 py-5">
            <span className="font-mono text-xs uppercase tracking-[0.12em] text-muted-foreground">
              Anatomy of a voice call
            </span>
            <span className="size-2 rounded-full bg-brand-orange" aria-hidden="true" />
          </div>
          <ol className="divide-y divide-border px-6">
            {callLayers.map((layer) => (
              <li key={layer.number} className="grid grid-cols-[2.5rem_1fr] gap-4 py-6">
                <span className="pt-1 font-mono text-xs tabular-nums text-brand-orange">
                  {layer.number}
                </span>
                <div>
                  <p className="font-mono text-[11px] uppercase tracking-[0.13em] text-muted-foreground">
                    {layer.label}
                  </p>
                  <p className="mt-2 font-heading text-xl font-medium tracking-tight">
                    {layer.title}
                  </p>
                  <p className="mt-1 text-sm leading-6 text-muted-foreground">
                    {layer.detail}
                  </p>
                </div>
              </li>
            ))}
          </ol>
          <div className="border-t border-border bg-background/40 px-6 py-4 font-mono text-xs text-muted-foreground">
            Media flows back to the caller through the SIP connection.
          </div>
        </div>
      </div>
    </section>
  );
}
