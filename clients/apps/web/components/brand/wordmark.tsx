export function Wordmark() {
  return (
    <section className="grid gap-4 sm:grid-cols-2">
      <div className="flex min-h-52 items-center justify-center rounded-xl border border-border bg-brand-offwhite text-brand-graphite">
        <span className="font-heading text-5xl font-semibold tracking-tight">
          Leamout<span className="text-brand-orange">.</span>
        </span>
      </div>
      <div className="flex min-h-52 items-center justify-center rounded-xl bg-brand-graphite text-brand-offwhite">
        <span className="font-heading text-5xl font-semibold tracking-tight">
          Leamout<span className="text-brand-orange">.</span>
        </span>
      </div>
    </section>
  );
}
