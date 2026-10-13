import { MarketingActions } from "@/components/actions";

export function HomeHero() {
  return (
    <section className="mx-auto grid max-w-6xl gap-14 px-5 py-20 sm:px-6 sm:py-28 lg:grid-cols-[1.2fr_1fr] lg:items-center lg:px-8">
      <div>
        <p className="mb-6 font-mono text-xs uppercase tracking-[0.18em] text-brand-orange">
          Your carriers. Your AI. Your control.
        </p>
        <h1 className="text-balance font-heading text-5xl font-semibold leading-[1.04] tracking-tight sm:text-6xl lg:text-7xl">
          Voice agents.
          <br />
          On your terms.
        </h1>
        <p className="mt-7 max-w-xl text-lg leading-8 text-muted-foreground">
          Leamout is the carrier-grade runtime and control plane for autonomous
          voice agents.
        </p>
        <p className="mb-8 mt-4 max-w-lg leading-7 text-muted-foreground">
          Connect your SIP infrastructure and AI providers. Build the
          conversation, connect your tools, and operate every call from one
          platform.
        </p>
        <MarketingActions />
      </div>
      <div className="rounded-2xl bg-brand-charcoal p-6 text-white sm:p-8">
        <p className="border-b border-white/20 pb-5 font-mono text-xs uppercase tracking-widest text-white/60">
          The voice agent stack
        </p>
        <div className="mt-7 grid grid-cols-2 gap-3">
          <div className="rounded-lg border border-white/20 p-4">
            <span className="font-mono text-xs text-white/60">
              01 / CONNECT
            </span>
            <p className="mt-3 font-medium">Your carrier</p>
            <p className="mt-1 text-xs text-white/60">SIP trunk · PBX · SBC</p>
          </div>
          <div className="rounded-lg border border-white/20 p-4">
            <span className="font-mono text-xs text-white/60">02 / CHOOSE</span>
            <p className="mt-3 font-medium">Your AI</p>
            <p className="mt-1 text-xs text-white/60">
              Realtime or STT + LLM + TTS
            </p>
          </div>
        </div>
        <div aria-hidden="true" className="mx-auto h-8 w-px bg-brand-orange" />
        <div className="rounded-lg bg-brand-orange p-6">
          <p className="font-heading text-2xl font-semibold">Leamout runtime</p>
          <p className="mt-2 text-sm leading-6">
            Conversation orchestration, call control, and tool execution.
          </p>
        </div>
        <div aria-hidden="true" className="mx-auto h-8 w-px bg-brand-orange" />
        <div className="flex items-center justify-between gap-3 rounded-lg border border-white/20 p-4">
          <span>Your application</span>
          <span className="font-mono text-xs text-white/60">
            API + WEBHOOKS
          </span>
        </div>
      </div>
    </section>
  );
}
