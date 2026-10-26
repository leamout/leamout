import { Button } from "@leamout/ui/components/button";
import Link from "next/link";
import { consoleURL } from "@/lib/site";
import { SessionPreview } from "./session-preview";

export function HomeHero() {
  return (
    <section className="relative overflow-hidden border-b border-border">
      <div className="mx-auto grid max-w-[1200px] gap-14 px-5 py-20 sm:px-6 sm:py-24 lg:grid-cols-[0.95fr_1.05fr] lg:items-center lg:gap-16 lg:px-8 lg:py-28">
        <div>
          <p className="mb-7 flex items-center gap-3 font-mono text-xs font-medium uppercase tracking-[0.16em] text-foreground">
            <span
              aria-hidden="true"
              className="size-2 rounded-full bg-brand-orange"
            />
            SIP-native voice infrastructure
          </p>
          <h1 className="text-balance font-heading text-5xl font-semibold leading-[1.06] tracking-tight sm:text-6xl xl:text-[5.25rem]">
            Voice agents.
            <br />
            <span className="text-muted-foreground">Built for real calls.</span>
          </h1>
          <p className="mt-8 max-w-xl text-lg leading-8 text-muted-foreground">
            Leamout is the carrier-grade runtime and control plane for
            autonomous voice agents. Connect your SIP trunks and AI providers,
            then build, run, and observe conversations in one managed service.
          </p>
          <div className="mt-9 flex flex-wrap gap-3">
            <Button
              size="lg"
              nativeButton={false}
              render={<a href={`${consoleURL}/create-account`} />}
            >
              Get started
            </Button>
            <Button
              size="lg"
              variant="outline"
              nativeButton={false}
              render={<Link href="/docs/quickstart" />}
            >
              Explore quickstart
            </Button>
          </div>
          <ul className="mt-12 flex flex-wrap gap-x-6 gap-y-3 border-t border-border pt-6 text-xs font-medium text-muted-foreground">
            <li>SIP-native</li>
            <li>Bring your carrier</li>
            <li>Choose your AI</li>
          </ul>
        </div>
        <SessionPreview />
      </div>
    </section>
  );
}
