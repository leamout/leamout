import Link from "next/link";

const steps = [
  {
    number: "01",
    title: "Connect your AI provider",
    description:
      "Add an organization-owned provider credential and verify the roles your agent will use.",
    href: "/docs/guides/ai/integrations",
    action: "AI integrations",
  },
  {
    number: "02",
    title: "Set up SIP connectivity",
    description:
      "Configure your carrier or PBX, add remote endpoints, and validate the trunk.",
    href: "/docs/guides/connectivity/sip-trunks",
    action: "SIP trunk setup",
  },
  {
    number: "03",
    title: "Configure the Voice Agent",
    description:
      "Define the instructions and engine, bind provider roles, and check readiness issues.",
    href: "/docs/guides/voice/voice-agents",
    action: "Voice Agents",
  },
  {
    number: "04",
    title: "Route an existing number",
    description:
      "Register a carrier-owned number, assign its trunk, and bind it to your agent for inbound calls.",
    href: "/docs/guides/voice/inbound-call",
    action: "Inbound calling",
  },
];

export function DeveloperQuickstart() {
  return (
    <section
      id="quickstart"
      aria-labelledby="developer-quickstart-title"
      className="border-b border-border bg-white"
    >
      <div className="mx-auto max-w-[1280px] px-5 py-20 sm:px-6 sm:py-24 lg:px-8 lg:py-28">
        <div className="grid gap-7 lg:grid-cols-[0.92fr_1.08fr] lg:items-end lg:gap-16">
          <div>
            <p className="font-mono text-xs uppercase tracking-[0.16em] text-brand-orange">
              04 / Developer onboarding
            </p>
            <h2
              id="developer-quickstart-title"
              className="mt-5 max-w-2xl text-balance font-heading text-4xl font-semibold leading-tight tracking-[-0.04em] sm:text-5xl"
            >
              From provider credentials to a call-ready agent.
            </h2>
          </div>
          <p className="max-w-2xl text-base leading-7 text-muted-foreground sm:text-lg sm:leading-8">
            Follow the current setup guides in order. Readiness checks the
            agent configuration; carrier ingress, inbound routing, and the
            complete call path still need to be verified separately.
          </p>
        </div>

        <ol className="mt-10 border-y border-border sm:mt-14">
          {steps.map((step) => (
            <li
              key={step.number}
              className="grid gap-3 border-b border-border py-6 last:border-b-0 sm:grid-cols-[48px_minmax(0,1fr)_auto] sm:items-center sm:gap-6 sm:py-7"
            >
              <span className="font-mono text-xs tracking-wider text-brand-orange">
                {step.number}
              </span>
              <div className="min-w-0">
                <h3 className="font-heading text-lg font-semibold tracking-tight">
                  {step.title}
                </h3>
                <p className="mt-2 max-w-2xl text-sm leading-6 text-muted-foreground">
                  {step.description}
                </p>
              </div>
              <Link
                href={step.href}
                className="mt-1 inline-flex items-center gap-2 self-start text-sm font-semibold underline decoration-brand-orange decoration-2 underline-offset-4 transition-colors hover:text-brand-orange sm:mt-0 sm:self-center"
              >
                {step.action}
                <span aria-hidden="true">↗</span>
              </Link>
            </li>
          ))}
        </ol>

        <div className="mt-10 grid gap-7 border border-border bg-background p-5 sm:p-7 lg:grid-cols-[0.8fr_1.2fr] lg:items-center lg:gap-10 lg:p-9">
          <div>
            <p className="font-mono text-[10px] uppercase tracking-[0.15em] text-brand-orange">
              Before you route live traffic
            </p>
            <h3 className="mt-4 font-heading text-2xl font-semibold tracking-tight">
              Check agent readiness.
            </h3>
            <p className="mt-3 text-sm leading-7 text-muted-foreground">
              Use an organization API token with permission and a saved Voice
              Agent ID. Inspect the ready value and address reported issues;
              this check does not validate the full carrier-to-model path.
            </p>
            <Link
              href="/docs/quickstart"
              className="mt-5 inline-flex items-center gap-2 text-sm font-semibold underline decoration-brand-orange decoration-2 underline-offset-4 transition-colors hover:text-brand-orange"
            >
              API quickstart
              <span aria-hidden="true">↗</span>
            </Link>
          </div>

          <div className="min-w-0 border border-foreground/15 bg-white">
            <div className="border-b border-border px-4 py-3 font-mono text-[10px] uppercase tracking-wider text-muted-foreground sm:px-5">
              GET /v1/voice-agents/:id/readiness
            </div>
            <pre className="overflow-x-auto p-4 font-mono text-[11px] leading-6 text-foreground sm:p-5 sm:text-xs"><code>{"curl --request GET --url https://api.leamout.com/v1/voice-agents/$VOICE_AGENT_ID/readiness --header 'Authorization: Bearer $LEAMOUT_API_TOKEN'"}</code></pre>
          </div>
        </div>
      </div>
    </section>
  );
}
