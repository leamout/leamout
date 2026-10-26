import { Button } from "@leamout/ui/components/button";
import Link from "next/link";
import { consoleURL } from "@/lib/site";

export function IntegrationsIntro() {
  return (
    <section className="mx-auto max-w-6xl px-5 py-20 sm:px-6 sm:py-28 lg:px-8">
      <div className="max-w-4xl">
        <p className="mb-6 font-mono text-xs uppercase tracking-[0.18em] text-brand-orange">
          Integrations
        </p>
        <h1 className="text-balance font-heading text-5xl font-semibold leading-[1.05] tracking-tight sm:text-6xl">
          Your carrier and AI stack.
          <br />
          Connected for voice.
        </h1>
        <p className="mt-7 max-w-2xl text-lg leading-8 text-muted-foreground">
          Set up existing SIP trunks and phone numbers, verify AI
          credentials, and bind application tools and webhooks. See how
          each connection fits into a live Voice Agent.
        </p>
        <div className="mt-8 flex flex-wrap gap-3">
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
            Integration documentation
          </Button>
        </div>
      </div>
    </section>
  );
}
