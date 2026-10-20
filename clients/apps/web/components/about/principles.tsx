const principles = [
  {
    number: "01",
    title: "Keep your connections",
    description:
      "Use your existing phone numbers and SIP carrier or PBX. Bring the AI providers that match your application.",
  },
  {
    number: "02",
    title: "Build with control",
    description:
      "Configure how agents respond, connect them to tools, and integrate events through APIs and webhooks.",
  },
  {
    number: "03",
    title: "Operate with context",
    description:
      "Manage calls and inspect conversation activity from the same control plane that runs your voice agents.",
  },
];

export function Principles() {
  return (
    <section className="border-b border-border py-20" aria-labelledby="about-principles">
      <div className="mb-10 flex flex-col justify-between gap-5 md:flex-row md:items-end">
        <div>
          <p className="font-mono text-xs uppercase tracking-[0.16em] text-brand-orange">
            Our approach
          </p>
          <h2
            id="about-principles"
            className="mt-5 max-w-xl text-balance font-heading text-3xl font-semibold tracking-tight sm:text-4xl"
          >
            Infrastructure that leaves you in control.
          </h2>
        </div>
        <p className="max-w-sm leading-7 text-muted-foreground">
          We build around three practical principles for running voice agents.
        </p>
      </div>
      <div className="border-t border-border">
        {principles.map((principle) => (
          <div
            key={principle.number}
            className="grid gap-3 border-b border-border py-7 md:grid-cols-[4rem_1fr_1.1fr] md:items-start md:gap-10"
          >
            <span className="font-mono text-xs tabular-nums text-brand-orange">
              {principle.number}
            </span>
            <h3 className="font-heading text-xl font-medium tracking-tight sm:text-2xl">
              {principle.title}
            </h3>
            <p className="max-w-lg leading-7 text-muted-foreground">
              {principle.description}
            </p>
          </div>
        ))}
      </div>
    </section>
  );
}
