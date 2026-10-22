export function Wordmark() {
  return (
    <section
      className="border-b border-border py-20"
      aria-labelledby="brand-wordmark-title"
    >
      <div className="mb-9 flex flex-col justify-between gap-4 md:flex-row md:items-end">
        <div>
          <p className="font-mono text-xs uppercase tracking-[0.16em] text-brand-orange">
            01 / Wordmark
          </p>
          <h2
            id="brand-wordmark-title"
            className="mt-5 font-heading text-3xl font-semibold tracking-tight sm:text-4xl"
          >
            Leamout, in one line.
          </h2>
        </div>
        <p className="max-w-sm leading-7 text-muted-foreground">
          The name stays legible on dark and light surfaces. The orange point is
          a restrained finishing detail.
        </p>
      </div>
      <div className="grid gap-px overflow-hidden rounded-xl border border-border bg-border md:grid-cols-2">
        <div className="flex min-h-64 flex-col justify-between bg-brand-offwhite p-7 text-brand-graphite sm:min-h-80 sm:p-9">
          <span className="font-mono text-xs uppercase tracking-wider text-brand-graphite/65">
            Light field
          </span>
          <p className="text-center font-heading text-5xl font-semibold tracking-tight sm:text-6xl">
            Leamout<span className="text-brand-orange">.</span>
          </p>
          <span className="font-mono text-xs tracking-wider text-brand-graphite/65">
            Graphite on Off-white
          </span>
        </div>
        <div className="flex min-h-64 flex-col justify-between bg-brand-graphite p-7 text-brand-offwhite sm:min-h-80 sm:p-9">
          <span className="font-mono text-xs uppercase tracking-wider text-brand-offwhite/60">
            Dark field
          </span>
          <p className="text-center font-heading text-5xl font-semibold tracking-tight sm:text-6xl">
            Leamout<span className="text-brand-orange">.</span>
          </p>
          <span className="font-mono text-xs tracking-wider text-brand-offwhite/60">
            Off-white on Graphite
          </span>
        </div>
      </div>
    </section>
  );
}
