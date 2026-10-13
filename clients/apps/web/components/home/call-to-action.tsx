import { MarketingActions } from "@/components/actions";

export function HomeCallToAction() {
  return (
    <section className="mx-auto max-w-6xl px-5 py-20 sm:px-6 lg:px-8">
      <h2 className="text-3xl font-semibold tracking-tight sm:text-4xl">
        Build your first voice agent.
      </h2>
      <p className="mb-7 mt-4 text-muted-foreground">
        Start with your own carrier and AI provider. Keep control as you grow.
      </p>
      <MarketingActions />
    </section>
  );
}
