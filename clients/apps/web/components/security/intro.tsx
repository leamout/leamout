const boundaries = [
  [
    "01",
    "Organization access",
    "Requests and resources are scoped to their organization.",
  ],
  ["02", "Provider secrets", "Credentials stay outside public API responses."],
  [
    "03",
    "Agent execution",
    "Tools and outbound webhooks require explicit controls.",
  ],
];

export function SecurityIntro() {
  return (
    <section className="grid gap-14 border-b border-border py-20 sm:py-24 lg:grid-cols-[1.1fr_0.9fr] lg:items-center lg:gap-16 lg:py-28">
      <div>
        <p className="mb-7 flex items-center gap-3 font-mono text-xs uppercase tracking-[0.16em]">
          <span
            className="size-2 rounded-full bg-brand-orange"
            aria-hidden="true"
          />
          Platform security
        </p>
        <h1 className="max-w-2xl text-balance font-heading text-5xl font-semibold leading-[1.06] tracking-tight sm:text-6xl">
          Security belongs in the voice path.
        </h1>
        <p className="mt-8 max-w-xl text-lg leading-8 text-muted-foreground">
          A voice agent crosses telephony, AI, and application boundaries.
          Leamout is designed to keep organizations isolated, credentials
          scoped, and agent actions under application control.
        </p>
        <p className="mt-5 max-w-xl leading-8 text-muted-foreground">
          This page describes the platform approach. It is not a claim of
          independent certification or a contractual security guarantee.
        </p>
      </div>
      <div className="overflow-hidden rounded-xl border border-border bg-brand-surface">
        <div className="flex items-center justify-between border-b border-border px-6 py-5">
          <p className="font-mono text-xs uppercase tracking-[0.14em] text-muted-foreground">
            Trust boundaries
          </p>
          <span
            className="size-2 rounded-full bg-brand-orange"
            aria-hidden="true"
          />
        </div>
        <ol className="divide-y divide-border px-6">
          {boundaries.map(([number, title, description]) => (
            <li key={number} className="grid grid-cols-[2.5rem_1fr] gap-4 py-6">
              <span className="font-mono text-xs text-brand-orange">
                {number}
              </span>
              <div>
                <p className="font-heading text-lg font-medium">{title}</p>
                <p className="mt-2 text-sm leading-6 text-muted-foreground">
                  {description}
                </p>
              </div>
            </li>
          ))}
        </ol>
      </div>
    </section>
  );
}
