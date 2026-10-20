import Link from "next/link";

export function Implementation() {
  return (
    <section className="my-16 rounded-xl bg-muted/40 p-8">
      <h2 className="text-2xl font-heading font-semibold">Review the implementation.</h2>
      <p className="my-4 max-w-2xl leading-7 text-muted-foreground">
        Evaluate the platform against your own requirements. This overview does
        not assert a compliance certification or contractual security guarantee.
      </p>
      <Link
        href="/docs"
        className="font-medium underline decoration-brand-orange underline-offset-8"
      >
        Explore the documentation
      </Link>
    </section>
  );
}
