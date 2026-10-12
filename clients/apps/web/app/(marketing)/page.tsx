import Link from "next/link";
import { MarketingActions } from "@/components/marketing/actions";
export default function Home() {
  return (
    <div>
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
            Leamout is the carrier-grade runtime and control plane for
            autonomous voice agents.
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
              <p className="mt-1 text-xs text-white/60">
                SIP trunk · PBX · SBC
              </p>
            </div>
            <div className="rounded-lg border border-white/20 p-4">
              <span className="font-mono text-xs text-white/60">
                02 / CHOOSE
              </span>
              <p className="mt-3 font-medium">Your AI</p>
              <p className="mt-1 text-xs text-white/60">
                Realtime or STT + LLM + TTS
              </p>
            </div>
          </div>
          <div
            aria-hidden="true"
            className="mx-auto h-8 w-px bg-brand-orange"
          />
          <div className="rounded-lg bg-brand-orange p-6">
            <p className="font-heading text-2xl font-semibold">
              Leamout runtime
            </p>
            <p className="mt-2 text-sm leading-6">
              Conversation orchestration, call control, and tool execution.
            </p>
          </div>
          <div
            aria-hidden="true"
            className="mx-auto h-8 w-px bg-brand-orange"
          />
          <div className="flex items-center justify-between gap-3 rounded-lg border border-white/20 p-4">
            <span>Your application</span>
            <span className="font-mono text-xs text-white/60">
              API + WEBHOOKS
            </span>
          </div>
        </div>
      </section>
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
      <section className="mx-auto max-w-6xl px-5 py-20 sm:px-6 lg:px-8">
        <h2 className="text-3xl font-semibold tracking-tight sm:text-4xl">
          The conversation is only the beginning.
        </h2>
        <div className="mt-10 grid gap-x-12 md:grid-cols-2">
          {[
            [
              "Call control",
              "Route calls, send DTMF, manage playback, and hand conversations off to people.",
            ],
            [
              "AI orchestration",
              "Compose speech recognition, language models, and synthesis, or connect an integrated realtime provider.",
            ],
            [
              "Tools and events",
              "Let agents act through configured tools. Connect application workflows through event-driven webhooks.",
            ],
            [
              "Operational visibility",
              "Review calls, recordings, and conversation history alongside the configuration that shaped each session.",
            ],
          ].map(([title, text]) => (
            <div key={title} className="border-t border-border py-7">
              <h3 className="text-xl font-semibold">{title}</h3>
              <p className="mt-3 leading-7 text-muted-foreground">{text}</p>
            </div>
          ))}
        </div>
      </section>
      <section className="bg-muted/30">
        <div className="mx-auto grid max-w-6xl gap-10 px-5 py-16 sm:px-6 md:grid-cols-2 lg:px-8">
          <h2 className="max-w-md text-3xl font-semibold leading-tight tracking-tight sm:text-4xl">
            Keep your provider relationships. Own your choices.
          </h2>
          <div>
            <p className="leading-8 text-muted-foreground">
              Leamout provides the runtime and control plane. Your carrier
              handles telephony, and your AI providers handle model usage. You
              pay those providers directly, with no Leamout carrier or AI resale
              charges.
            </p>
            <Link
              href="/pricing"
              className="mt-6 inline-block font-medium underline decoration-brand-orange underline-offset-8"
            >
              Explore platform plans
            </Link>
          </div>
        </div>
      </section>
      <section className="mx-auto max-w-6xl px-5 py-20 sm:px-6 lg:px-8">
        <h2 className="text-3xl font-semibold tracking-tight sm:text-4xl">
          Build your first voice agent.
        </h2>
        <p className="mb-7 mt-4 text-muted-foreground">
          Start with your own carrier and AI provider. Keep control as you grow.
        </p>
        <MarketingActions />
      </section>
    </div>
  );
}
