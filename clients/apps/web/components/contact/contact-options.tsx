import { Button } from "@leamout/ui/components/button";
import Link from "next/link";
import { consoleURL } from "@/lib/site";

const issueURL = "https://github.com/leamout/leamout/issues/new";

export function ContactOptions() {
  return (
    <div className="py-20">
      <div className="grid gap-px overflow-hidden rounded-xl border border-border bg-border md:grid-cols-2 lg:grid-cols-3">
        <section className="flex flex-col bg-card p-8 sm:p-10">
          <p className="font-mono text-xs uppercase tracking-[0.16em] text-brand-orange">
            01 / Get building
          </p>
          <h2 className="mt-6 font-heading text-3xl font-semibold tracking-tight">
            Explore Leamout
          </h2>
          <p className="mt-4 max-w-md flex-1 leading-7 text-muted-foreground">
            Start with the documentation to understand SIP trunks, AI provider
            connections, and the agent runtime. Create an organization when you
            are ready to build.
          </p>
          <div className="mt-9 flex flex-wrap gap-3">
            <Button
              nativeButton={false}
              render={<a href={`${consoleURL}/create-account`} />}
            >
              Get started
            </Button>
            <Button
              variant="outline"
              nativeButton={false}
              render={<Link href="/docs" />}
            >
              Read the docs
            </Button>
          </div>
        </section>
        <section className="flex flex-col bg-card p-8 sm:p-10">
          <p className="font-mono text-xs uppercase tracking-[0.16em] text-brand-orange">
            02 / Private inquiries
          </p>
          <h2 className="mt-6 font-heading text-3xl font-semibold tracking-tight">
            Contact the team
          </h2>
          <p className="mt-4 max-w-md flex-1 leading-7 text-muted-foreground">
            For business questions, partnerships, or account-specific help,
            reach us privately. Please avoid including credentials or
            recordings in your initial message.
          </p>
          <div className="mt-9">
            <Button
              variant="outline"
              nativeButton={false}
              render={<a href="mailto:hello@leamout.com" />}
            >
              hello@leamout.com
            </Button>
          </div>
        </section>
        <section className="flex flex-col bg-card p-8 sm:p-10">
          <p className="font-mono text-xs uppercase tracking-[0.16em] text-brand-orange">
            03 / Feedback
          </p>
          <h2 className="mt-6 font-heading text-3xl font-semibold tracking-tight">
            Share a product issue
          </h2>
          <p className="mt-4 max-w-md flex-1 leading-7 text-muted-foreground">
            For non-sensitive bugs or feature requests, use the public issue
            tracker. Include a clear description and reproduction steps, but
            never post API keys, call recordings, customer information, or
            provider credentials.
          </p>
          <div className="mt-9">
            <Button
              variant="outline"
              nativeButton={false}
              render={
                <a href={issueURL} target="_blank" rel="noopener noreferrer" />
              }
            >
              Open a GitHub issue
            </Button>
          </div>
        </section>
      </div>

      <section className="mt-6 flex flex-col justify-between gap-6 rounded-xl border border-border bg-brand-surface p-8 sm:p-10 lg:flex-row lg:items-center">
        <div>
          <p className="font-mono text-xs uppercase tracking-[0.16em] text-brand-orange">
            Private security reports
          </p>
          <h2 className="mt-3 font-heading text-2xl font-semibold tracking-tight">
            Found a potential vulnerability?
          </h2>
          <p className="mt-3 max-w-2xl leading-7 text-muted-foreground">
            Please report vulnerabilities privately. Do not disclose security
            issues through a public issue or discussion.
          </p>
        </div>
        <a
          className="shrink-0 font-medium text-foreground underline decoration-brand-orange underline-offset-8 transition-colors hover:text-brand-orange"
          href="mailto:security@leamout.com"
        >
          security@leamout.com
        </a>
      </section>
    </div>
  );
}
