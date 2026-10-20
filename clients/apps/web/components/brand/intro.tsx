export function BrandIntro() {
  return (
    <section className="grid gap-8 border-b border-border py-20 sm:py-24 lg:grid-cols-[1.15fr_0.85fr] lg:items-end lg:gap-16 lg:py-28">
      <div>
        <p className="mb-7 flex items-center gap-3 font-mono text-xs uppercase tracking-[0.16em]">
          <span className="size-2 rounded-full bg-brand-orange" aria-hidden="true" />
          Leamout identity
        </p>
        <h1 className="max-w-3xl text-balance font-heading text-5xl font-semibold leading-[1.06] tracking-tight sm:text-6xl">
          A clear identity for serious infrastructure.
        </h1>
      </div>
      <p className="max-w-md text-lg leading-8 text-muted-foreground">
        Graphite foundations, off-white typography, and one decisive orange.
        The same visual language runs across the Leamout website and console.
      </p>
    </section>
  );
}
