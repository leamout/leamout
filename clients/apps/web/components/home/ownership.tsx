import Link from "next/link";

export function Ownership() {
  return (
    <section className="bg-muted/30">
      <div className="mx-auto grid max-w-6xl gap-10 px-5 py-16 sm:px-6 md:grid-cols-2 lg:px-8">
        <h2 className="max-w-md text-3xl font-semibold leading-tight tracking-tight sm:text-4xl">
          Keep your provider relationships. Own your choices.
        </h2>
        <div>
          <p className="leading-8 text-muted-foreground">
            Leamout operates the runtime and control plane. Your carrier handles
            telephony, and your AI providers handle model usage. You pay those
            providers directly, with no Leamout carrier or AI resale charges.
          </p>
          <Link
            href="/pricing"
            className="mt-6 inline-block font-medium underline decoration-brand-orange underline-offset-8"
          >
            Explore platform plans
          </Link>
        </div>
      </div>
    </section>
  );
}
