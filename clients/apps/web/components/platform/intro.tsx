import { Button } from "@leamout/ui/components/button";
import Link from "next/link";
import { consoleURL } from "@/lib/site";

export function PlatformIntro() {
  return (
    <section className="mx-auto max-w-6xl px-5 py-20 sm:px-6 sm:py-28 lg:px-8">
      <div className="max-w-4xl">
        <p className="mb-6 font-mono text-xs uppercase tracking-[0.18em] text-brand-orange">
          The Leamout platform
        </p>
        <h1 className="text-balance font-heading text-5xl font-semibold leading-[1.05] tracking-tight sm:text-6xl">
          The runtime behind every conversation.
        </h1>
        <p className="mt-7 max-w-2xl text-lg leading-8 text-muted-foreground">
          Leamout brings SIP calling, realtime media, voice agent orchestration,
          and operational control together in one managed service.
        </p>
        <p className="mt-4 max-w-2xl leading-8 text-muted-foreground">
          Connect your carrier and AI providers. Leamout runs the conversation
          while you define how your agents behave and what they can do.
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
            Explore documentation
          </Button>
        </div>
      </div>
    </section>
  );
}
