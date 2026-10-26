import Link from "next/link";

const setup = [
  {
    title: "Configure your trunk",
    description:
      "Register the carrier connection, permitted source networks, and any required authentication.",
  },
  {
    title: "Validate SIP connectivity",
    description:
      "Add remote endpoints where needed and review trunk configuration before directing live traffic.",
  },
  {
    title: "Route an existing number",
    description:
      "Register your carrier-owned number, assign the trunk, and bind the number to a ready Voice Agent.",
  },
];

export function IntegrationsTelephony() {
  return (
    <section className="theme-light border-y border-border bg-background text-foreground">
      <div className="mx-auto grid max-w-6xl gap-10 px-5 py-20 sm:px-6 lg:grid-cols-[0.85fr_1.15fr] lg:gap-16 lg:px-8 lg:py-28">
        <div>
          <p className="font-mono text-xs uppercase tracking-widest text-foreground">
            Bring your own carrier
          </p>
          <h2 className="mt-4 text-balance font-heading text-3xl font-semibold tracking-tight sm:text-4xl">
            Connect the SIP service you already operate.
          </h2>
          <p className="mt-6 leading-8 text-muted-foreground">
            Keep your carrier account, numbers, and telephony billing. Leamout
            handles the managed voice runtime around that connection; it
            does not sell or provision phone numbers.
          </p>
          <Link
            href="/docs/guides/connectivity/sip-trunks"
            className="mt-7 inline-block text-sm font-medium text-foreground underline decoration-brand-orange underline-offset-4 hover:text-foreground"
          >
            Configure a SIP trunk
          </Link>
        </div>
        <div className="self-center overflow-hidden rounded-xl border border-border bg-card">
          <div className="border-b border-border px-6 py-5">
            <p className="font-mono text-xs uppercase tracking-widest text-muted-foreground">
              Connection checklist
            </p>
          </div>
          <ol className="divide-y divide-border">
            {setup.map((item, index) => (
              <li key={item.title} className="flex gap-5 px-6 py-6">
                <span className="shrink-0 font-mono text-xs text-brand-orange">
                  {String(index + 1).padStart(2, "0")}
                </span>
                <div>
                  <h3 className="font-semibold">{item.title}</h3>
                  <p className="mt-2 text-sm leading-7 text-muted-foreground">
                    {item.description}
                  </p>
                </div>
              </li>
            ))}
          </ol>
        </div>
      </div>
    </section>
  );
}
