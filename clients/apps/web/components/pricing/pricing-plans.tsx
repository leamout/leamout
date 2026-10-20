import { Button } from "@leamout/ui/components/button";
import { cn } from "@leamout/ui/lib/utils";
import Link from "next/link";
import { plans } from "@/lib/plans";

const presentation = {
  Free: {
    eyebrow: "Explore",
    description: "Build and test your first voice agent.",
  },
  Developer: {
    eyebrow: "Launch",
    description: "Move your first voice applications into production.",
  },
  Pro: {
    eyebrow: "Operate",
    description: "Run more calls with expanded operational controls.",
  },
  Scale: {
    eyebrow: "Expand",
    description: "Support higher-volume agents and larger teams.",
  },
};

export default function PricingPlans({ signupHref }: { signupHref: string }) {
  return (
    <section
      className="border-b border-border bg-background py-20 sm:py-24 lg:py-28"
      aria-labelledby="pricing-title"
    >
      <div className="mx-auto max-w-[1200px] px-5 sm:px-6 lg:px-8">
        <div className="mb-12 grid gap-7 lg:grid-cols-[1.2fr_0.8fr] lg:items-end">
          <div>
            <p className="mb-6 flex items-center gap-3 font-mono text-xs uppercase tracking-[0.16em]">
              <span className="size-2 rounded-full bg-brand-orange" aria-hidden="true" />
              Straightforward pricing
            </p>
            <h1
              id="pricing-title"
              className="max-w-3xl text-balance font-heading text-5xl font-semibold leading-[1.06] tracking-tight sm:text-6xl"
            >
              One runtime.
              <br />
              <span className="text-muted-foreground">Room to grow.</span>
            </h1>
          </div>
          <p className="max-w-md text-lg leading-8 text-muted-foreground lg:justify-self-end">
            Choose the capacity you need to run autonomous voice agents.
            Leamout manages the runtime; you connect your SIP carrier and AI
            providers.
          </p>
        </div>

        <div className="grid gap-px overflow-hidden rounded-xl border border-border bg-border sm:grid-cols-2 xl:grid-cols-4">
          {plans.map((plan) => {
            const details =
              presentation[plan.name as keyof typeof presentation];
            const highlighted = plan.name === "Pro";

            return (
              <article
                key={plan.name}
                className={cn(
                  "relative flex min-w-0 flex-col bg-card p-6 sm:p-7 xl:p-6",
                  highlighted && "bg-brand-surface",
                )}
              >
                {highlighted && (
                  <span
                    aria-hidden="true"
                    className="absolute inset-x-0 top-0 h-1 bg-brand-orange"
                  />
                )}
                <div className="min-h-32">
                  <p className="font-mono text-xs uppercase tracking-[0.14em] text-brand-orange">
                    {details.eyebrow}
                  </p>
                  <h2 className="mt-3 font-heading text-2xl font-semibold tracking-tight">
                    {plan.name}
                  </h2>
                  <p className="mt-3 max-w-xs text-sm leading-6 text-muted-foreground">
                    {details.description}
                  </p>
                </div>

                <div className="mt-7">
                  <p className="flex flex-wrap items-baseline gap-x-1.5">
                    <span className="font-heading text-5xl font-semibold tabular-nums tracking-tight">
                      {"$" + plan.price.toLocaleString("en-US")}
                    </span>
                    <span className="text-sm text-muted-foreground">
                      / month
                    </span>
                  </p>
                  <p className="mt-2 text-xs text-muted-foreground">
                    USD, billed monthly
                  </p>
                </div>

                <Button
                  className="mt-7 w-full"
                  variant={highlighted ? "default" : "outline"}
                  nativeButton={false}
                  render={
                    <a
                      href={signupHref}
                      aria-label={"Get started with " + plan.name}
                    />
                  }
                >
                  Get started
                </Button>

                <div className="mt-9 flex-1 border-t border-border pt-6">
                  <p className="mb-5 font-mono text-[11px] uppercase tracking-[0.12em] text-muted-foreground">
                    Monthly capacity
                  </p>
                  <p className="font-heading text-3xl font-semibold tabular-nums tracking-tight">
                    {plan.minutes.toLocaleString("en-US")}
                    <span className="ml-2 font-sans text-sm font-normal text-muted-foreground">
                      runtime minutes
                    </span>
                  </p>
                  <dl className="mt-6 space-y-3">
                    {[
                      ["Concurrent calls", plan.concurrency.toLocaleString("en-US")],
                      ["Voice agents", plan.agents.toLocaleString("en-US")],
                      ["SIP trunks", plan.trunks.toLocaleString("en-US")],
                      ["Retention", plan.retention + (plan.retention === 1 ? " day" : " days")],
                    ].map(([label, value]) => (
                      <div
                        key={label}
                        className="flex items-baseline justify-between gap-3 text-sm leading-6"
                      >
                        <dt className="text-muted-foreground">{label}</dt>
                        <dd className="font-medium tabular-nums">{value}</dd>
                      </div>
                    ))}
                  </dl>
                </div>

                <p className="mt-7 min-h-12 border-t border-border pt-5 text-sm leading-6 text-muted-foreground">
                  {highlighted || plan.name === "Scale"
                    ? "Includes advanced roles and configurable retention policies."
                    : "Includes access to the managed voice agent runtime."}
                </p>
              </article>
            );
          })}
        </div>

        <div className="mt-6 flex flex-col justify-between gap-5 text-sm text-muted-foreground sm:flex-row sm:items-center">
          <p>
            Every plan runs on the same managed Leamout platform. Carrier and
            AI usage charges are separate where applicable.
          </p>
          <Link
            href="/contact"
            className="shrink-0 font-medium text-foreground underline underline-offset-4 transition-colors hover:text-brand-orange"
          >
            Need custom capacity? Contact us
          </Link>
        </div>
      </div>
    </section>
  );
}
