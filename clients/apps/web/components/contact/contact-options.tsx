import Link from "next/link";

export function ContactOptions() {
  return (
    <div className="grid gap-6 pb-20 md:grid-cols-2">
      <section className="rounded-xl border border-border p-8">
        <h2 className="text-2xl font-heading font-semibold">Building with Leamout?</h2>
        <p className="my-5 leading-7 text-muted-foreground">
          Start with the Leamout console and product documentation.
          Include the affected organization, call timestamp and non-sensitive
          diagnostic details when you contact support.
        </p>
        <Link
          href="/docs"
          className="font-medium underline decoration-brand-orange underline-offset-8"
        >
          Read the documentation
        </Link>
      </section>
      <section className="rounded-xl border border-border p-8">
        <h2 className="text-2xl font-heading font-semibold">Questions and feedback</h2>
        <p className="my-5 leading-7 text-muted-foreground">
          Questions about Leamout should go through the documented
          support channels. Never publish credentials, recordings, personal
          information or vulnerability reports in a public issue.
        </p>
        <Link
          href="/contact"
          className="font-medium underline decoration-brand-orange underline-offset-8"
        >
          Cloud contact information
        </Link>
      </section>
    </div>
  );
}
