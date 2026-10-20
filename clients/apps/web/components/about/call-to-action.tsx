import { Button } from "@leamout/ui/components/button";
import Link from "next/link";
import { consoleURL } from "@/lib/site";

export function AboutCallToAction() {
  return (
    <section className="py-20">
      <div className="flex flex-col justify-between gap-10 rounded-xl border border-border bg-brand-surface p-8 sm:p-10 lg:flex-row lg:items-end lg:p-12">
        <div>
          <p className="font-mono text-xs uppercase tracking-[0.16em] text-brand-orange">
            Build on Leamout
          </p>
          <h2 className="mt-5 max-w-xl text-balance font-heading text-3xl font-semibold tracking-tight sm:text-4xl">
            Ready to build an agent that can take a call?
          </h2>
          <p className="mt-4 max-w-lg leading-7 text-muted-foreground">
            Connect your SIP carrier, configure your AI providers, and start
            building with the runtime.
          </p>
        </div>
        <div className="flex shrink-0 flex-wrap gap-3">
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
            Read the docs
          </Button>
        </div>
      </div>
    </section>
  );
}
