import { Button } from "@leamout/ui/components/button";
import Link from "next/link";
import { consoleURL } from "@/lib/site";
import { VoicePathVisual } from "./voice-path-visual";

export function HomeHero() {
  return (
    <section className="relative isolate overflow-hidden border-b border-border bg-background">
      <div
        aria-hidden="true"
        className="pointer-events-none absolute inset-0 -z-10"
        style={{
          backgroundImage:
            "linear-gradient(to right, rgba(16,19,22,0.045) 1px, transparent 1px), linear-gradient(to bottom, rgba(16,19,22,0.045) 1px, transparent 1px)",
          backgroundSize: "36px 36px",
          maskImage:
            "linear-gradient(to bottom, black 0%, black 70%, transparent 100%)",
        }}
      />
      <div className="mx-auto max-w-[1280px] px-5 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between gap-4 border-y border-border py-3 font-mono text-[10px] uppercase tracking-[0.14em] text-muted-foreground sm:text-[11px]">
          <span className="flex items-center gap-2">
            <span aria-hidden="true" className="size-2 bg-brand-orange" />
            Leamout / Voice infrastructure
          </span>
          <span className="hidden sm:inline">SIP-native runtime · 001</span>
        </div>

        <div className="grid gap-12 py-14 sm:py-16 lg:grid-cols-[0.92fr_1.08fr] lg:items-center lg:gap-12 lg:py-20 xl:gap-16">
          <div className="max-w-2xl">
            <p className="mb-6 font-mono text-xs font-medium uppercase tracking-[0.18em] text-brand-orange">
              Infrastructure for real conversations
            </p>
            <h1 className="text-balance font-heading text-5xl font-semibold leading-[0.96] tracking-[-0.055em] sm:text-6xl lg:text-7xl xl:text-8xl">
              Build the agent.
              <span className="mt-2 block text-muted-foreground">
                We run the call.
              </span>
            </h1>
            <p className="mt-7 max-w-xl text-base leading-7 text-muted-foreground sm:text-lg sm:leading-8">
              Leamout is the SIP-native voice runtime between your telephony and
              AI stack. We handle signaling, realtime media, agent execution,
              and call lifecycle operations so you can focus on the experience
              you are building.
            </p>
            <div className="mt-8 flex flex-wrap gap-3">
              <Button
                size="lg"
                nativeButton={false}
                render={<a href={`${consoleURL}/create-account`} />}
              >
                Start building
              </Button>
              <Button
                size="lg"
                variant="outline"
                nativeButton={false}
                render={<Link href="/docs/quickstart" />}
              >
                Read the quickstart
              </Button>
            </div>

            <div className="mt-11 grid max-w-xl grid-cols-3 border-t border-border pt-5">
              <div className="pr-3">
                <p className="font-mono text-[10px] uppercase tracking-wider text-brand-orange">
                  01
                </p>
                <p className="mt-2 text-xs font-medium sm:text-sm">
                  Bring your SIP
                </p>
              </div>
              <div className="border-l border-border px-3 sm:px-4">
                <p className="font-mono text-[10px] uppercase tracking-wider text-brand-orange">
                  02
                </p>
                <p className="mt-2 text-xs font-medium sm:text-sm">
                  Choose your AI
                </p>
              </div>
              <div className="border-l border-border pl-3 sm:pl-4">
                <p className="font-mono text-[10px] uppercase tracking-wider text-brand-orange">
                  03
                </p>
                <p className="mt-2 text-xs font-medium sm:text-sm">
                  Run the session
                </p>
              </div>
            </div>
          </div>

          <div className="min-w-0 lg:pl-2">
            <VoicePathVisual />
          </div>
        </div>
      </div>
    </section>
  );
}
