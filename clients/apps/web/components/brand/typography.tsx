const families = [
  {
    number: "01",
    name: "Space Grotesk",
    role: "Headlines and display",
    sample: "Built for real calls.",
    className: "font-heading text-3xl font-semibold tracking-tight sm:text-4xl",
  },
  {
    number: "02",
    name: "IBM Plex Sans",
    role: "Interfaces and body copy",
    sample: "Infrastructure developers can work with.",
    className: "font-sans text-2xl sm:text-3xl",
  },
  {
    number: "03",
    name: "IBM Plex Mono",
    role: "Labels and technical details",
    sample: "SIP / RTP / AGENT RUNTIME",
    className: "font-mono text-sm tracking-[0.12em] sm:text-base",
  },
];

export function Typography() {
  return (
    <section className="py-20" aria-labelledby="brand-type-title">
      <div className="mb-10">
        <p className="font-mono text-xs uppercase tracking-[0.16em] text-brand-orange">
          03 / Typography
        </p>
        <h2
          id="brand-type-title"
          className="mt-5 font-heading text-3xl font-semibold tracking-tight sm:text-4xl"
        >
          Technical clarity, human voice.
        </h2>
      </div>
      <div className="border-t border-border">
        {families.map((family) => (
          <div
            key={family.number}
            className="grid gap-6 border-b border-border py-8 lg:grid-cols-[3rem_0.8fr_1.5fr] lg:gap-10"
          >
            <span className="font-mono text-xs text-brand-orange">
              {family.number}
            </span>
            <div>
              <h3 className="font-heading text-xl font-medium">
                {family.name}
              </h3>
              <p className="mt-2 text-sm text-muted-foreground">
                {family.role}
              </p>
            </div>
            <p className={family.className}>{family.sample}</p>
          </div>
        ))}
      </div>
      <div className="mt-10 max-w-2xl">
        <h3 className="font-heading text-xl font-semibold">Using the name</h3>
        <p className="mt-3 leading-7 text-muted-foreground">
          Write “Leamout” with an uppercase L. Do not imply an endorsement,
          partnership, or certification without permission. Keep clear space
          around the wordmark so it remains readable.
        </p>
      </div>
    </section>
  );
}
