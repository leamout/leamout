import { Button } from "@leamout/ui/components/button";
import Link from "next/link";
import { consoleURL } from "@/lib/site";

export function HomeCallToAction() {
  return (
    <section className="border-b border-border bg-brand-orange text-brand-graphite">
      <div className="mx-auto max-w-[1280px] px-5 py-14 sm:px-6 sm:py-16 lg:px-8 lg:py-20">
        <div className="grid gap-8 lg:grid-cols-[1fr_auto] lg:items-end lg:gap-16">
          <div className="max-w-3xl">
            <p className="font-mono text-[10px] font-medium uppercase tracking-[0.16em]">
              05 / Start with one call
            </p>
            <h2 className="mt-5 text-balance font-heading text-4xl font-semibold leading-tight tracking-[-0.04em] sm:text-5xl lg:text-6xl">
              Make the next integration a real conversation.
            </h2>
            <p className="mt-5 max-w-2xl text-base leading-7 sm:text-lg sm:leading-8">
              Bring your existing SIP connectivity and AI provider. Start with
              the documented setup flow, then verify your route before
              sending real traffic.
            </p>
          </div>
          <div className="flex flex-wrap gap-3 lg:justify-end">
            <Button
              size="lg"
              nativeButton={false}
              className="bg-brand-graphite text-brand-offwhite hover:bg-brand-surface"
              render={<a href={`${consoleURL}/create-account`} />}
            >
              Get started
            </Button>
            <Button
              size="lg"
              variant="outline"
              className="border-brand-graphite/50 bg-transparent text-brand-graphite hover:bg-brand-graphite/10"
              nativeButton={false}
              render={<Link href="/docs/quickstart" />}
            >
              Read the quickstart
            </Button>
          </div>
        </div>
      </div>
    </section>
  );
}
