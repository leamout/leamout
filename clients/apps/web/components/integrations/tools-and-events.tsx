import Link from "next/link";

export function ToolsAndEvents() {
  return (
    <section className="theme-light border-t border-border bg-background text-foreground">
      <div className="mx-auto max-w-6xl px-5 py-20 sm:px-6 lg:px-8">
        <p className="font-mono text-xs uppercase tracking-widest text-foreground">
          Your applications
        </p>
        <h2 className="mt-4 max-w-2xl text-balance text-3xl font-heading font-semibold tracking-tight sm:text-4xl">
          Connect the conversation to your workflows.
        </h2>
        <div className="mt-12 grid gap-8 md:grid-cols-2">
          <div className="border-t border-border pt-6">
            <h3 className="text-xl font-semibold">Agent tools</h3>
            <p className="mt-4 leading-8 text-muted-foreground">
              Give agents access to configured tools so a conversation can
              trigger actions in your application. Control which tools are
              available to each agent.
            </p>
          </div>
          <div className="border-t border-border pt-6">
            <h3 className="text-xl font-semibold">Webhooks and call events</h3>
            <p className="mt-4 leading-8 text-muted-foreground">
              Receive events from Leamout and connect call activity to your
              own systems, monitoring, and operational workflows.
            </p>
          </div>
        </div>
        <Link
          href="/docs"
          className="mt-10 inline-block font-medium underline decoration-brand-orange underline-offset-8"
        >
          Explore integration documentation
        </Link>
      </div>
    </section>
  );
}
