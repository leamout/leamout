export function HowItWorks() {
  return (
    <section className="border-y border-border bg-muted/20">
      <div className="mx-auto max-w-6xl px-5 py-16 sm:px-6 lg:px-8">
        <p className="font-mono text-xs uppercase tracking-widest text-brand-orange">
          From connection to conversation
        </p>
        <h2 className="mt-4 max-w-xl text-3xl font-semibold tracking-tight sm:text-4xl">
          Build on the services you already trust.
        </h2>
        <ol className="mt-12 grid gap-8 md:grid-cols-3">
          {[
            [
              "01",
              "Connect your infrastructure",
              "Bring your existing numbers through a SIP trunk. Connect organization-owned AI provider credentials.",
            ],
            [
              "02",
              "Define your agent",
              "Set instructions, select a realtime model or speech pipeline, and connect tools to your application.",
            ],
            [
              "03",
              "Run and observe",
              "Start calls, inspect conversation activity, manage recordings, and receive events through webhooks.",
            ],
          ].map(([number, title, description]) => (
            <li key={number} className="border-t border-border pt-5">
              <span className="font-mono text-sm text-brand-orange">
                {number}
              </span>
              <h3 className="mt-4 text-lg font-semibold">{title}</h3>
              <p className="mt-3 leading-7 text-muted-foreground">
                {description}
              </p>
            </li>
          ))}
        </ol>
      </div>
    </section>
  );
}
