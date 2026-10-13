import { Button } from "@leamout/ui/components/button";
import Link from "next/link";
import { consoleURL } from "@/lib/site";

export function HomeHero() {
  return (
    <section className="mx-auto max-w-6xl px-5 py-20 sm:px-6 sm:py-28 lg:px-8">
      <div className="max-w-3xl">
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
        <div className="flex flex-wrap gap-3">
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
            render={<Link href="/docs" />}
          >
            Read documentation
          </Button>
        </div>
      </div>
    </section>
  );
}
