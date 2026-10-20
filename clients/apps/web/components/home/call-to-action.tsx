import { Button } from "@leamout/ui/components/button";
import Link from "next/link";
import { consoleURL } from "@/lib/site";

export function HomeCallToAction() {
  return (
    <section className="mx-auto max-w-[1200px] px-5 py-24 sm:px-6 lg:px-8 lg:py-28">
      <div className="relative overflow-hidden rounded-2xl border border-border bg-card px-7 py-14 sm:px-12 sm:py-16">
        <span
          aria-hidden="true"
          className="absolute inset-x-0 top-0 h-1 bg-brand-orange"
        />
        <p className="font-mono text-xs uppercase tracking-[0.16em] text-brand-orange">
          Start building
        </p>
        <h2 className="mt-5 max-w-2xl text-balance text-4xl font-heading font-semibold tracking-tight sm:text-5xl">
          Put your next voice agent on the line.
        </h2>
        <p className="mt-5 max-w-xl leading-8 text-muted-foreground">
          Connect your SIP trunk and AI providers. Define your agent, then let
          Leamout manage the runtime behind every conversation.
        </p>
        <div className="mt-8 flex flex-wrap gap-3">
          <Button
            size="lg"
            nativeButton={false}
            render={<a href={consoleURL + "/create-account"} />}
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
