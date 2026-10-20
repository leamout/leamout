const costLayers = [
  {
    number: "01",
    title: "Leamout",
    responsibility: "Platform subscription",
    description:
      "Your monthly plan covers the listed runtime allowance, call concurrency, voice agents, SIP trunks, and retention. Pro and Scale include additional control features.",
  },
  {
    number: "02",
    title: "Your SIP carrier",
    responsibility: "Telephony charges",
    description:
      "Bring your own carrier and phone numbers. Your carrier bills you directly for numbers and telephony usage; Leamout does not include carrier minutes.",
  },
  {
    number: "03",
    title: "Your AI providers",
    responsibility: "Model and speech usage",
    description:
      "Bring your own AI credentials and pay those providers directly. If Leamout-provided AI is enabled for your organization, its usage may be charged separately.",
  },
];

export function BillingExplanation() {
  return (
    <section className="py-20" aria-labelledby="pricing-billing-title">
      <div className="mb-10 grid gap-6 md:grid-cols-2 md:items-end">
        <div>
          <p className="font-mono text-xs uppercase tracking-[0.16em] text-brand-orange">
            How billing works
          </p>
          <h2
            id="pricing-billing-title"
            className="mt-5 max-w-lg text-balance font-heading text-3xl font-semibold leading-tight tracking-tight sm:text-4xl"
          >
            The platform fee is one part of the picture.
          </h2>
        </div>
        <p className="max-w-md leading-7 text-muted-foreground">
          You can keep your existing telephony and AI relationships. Here is
          what each part of your voice stack pays for.
        </p>
      </div>
      <div className="border-t border-border">
        {costLayers.map((layer) => (
          <div
            key={layer.number}
            className="grid gap-3 border-b border-border py-7 md:grid-cols-[3.5rem_1fr_1fr_1.6fr] md:gap-6"
          >
            <span className="font-mono text-xs tabular-nums text-brand-orange">
              {layer.number}
            </span>
            <h3 className="font-heading text-xl font-medium tracking-tight">
              {layer.title}
            </h3>
            <p className="text-sm font-medium text-foreground">
              {layer.responsibility}
            </p>
            <p className="max-w-lg text-sm leading-7 text-muted-foreground">
              {layer.description}
            </p>
          </div>
        ))}
      </div>
    </section>
  );
}
