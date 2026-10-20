import Link from "next/link";

export function Ownership() {
  return (
    <section className="marketing-light border-b border-border bg-background text-foreground">
      <div className="mx-auto max-w-[1200px] px-5 py-24 sm:px-6 lg:px-8 lg:py-28">
        <div className="grid gap-10 rounded-2xl border border-border bg-card p-7 sm:p-10 lg:grid-cols-[0.9fr_1.1fr] lg:gap-16 lg:p-14">
          <div>
            <p className="font-mono text-xs uppercase tracking-[0.16em] text-foreground">
              Your infrastructure, connected
            </p>
            <h2 className="mt-5 text-balance text-4xl font-semibold tracking-tight sm:text-5xl">
              Keep your providers. Run on Leamout.
            </h2>
            <p className="mt-6 leading-8 text-muted-foreground">
              Leamout operates the voice agent runtime and control plane. Your
              team keeps its existing carrier and AI provider relationships.
            </p>
            <Link
              href="/integrations"
              className="mt-8 inline-block font-medium text-foreground underline decoration-brand-orange underline-offset-8 hover:text-foreground"
            >
              Explore integrations
            </Link>
          </div>
          <div className="grid gap-4 self-center sm:grid-cols-2">
            <div className="border-t-2 border-brand-orange bg-background p-6">
              <p className="font-mono text-[11px] uppercase tracking-widest text-muted-foreground">
                Telephony
              </p>
              <h3 className="mt-7 text-xl font-semibold">Your SIP trunk</h3>
              <p className="mt-3 text-sm leading-7 text-muted-foreground">
                Bring your carrier, phone numbers, and SIP connectivity.
              </p>
            </div>
            <div className="border-t-2 border-brand-orange bg-background p-6">
              <p className="font-mono text-[11px] uppercase tracking-widest text-muted-foreground">
                Intelligence
              </p>
              <h3 className="mt-7 text-xl font-semibold">Your AI stack</h3>
              <p className="mt-3 text-sm leading-7 text-muted-foreground">
                Select a realtime model or compose STT, LLM, and TTS providers.
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
