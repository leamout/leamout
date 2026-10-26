import Link from "next/link";

const steps = [
  {
    number: "01",
    title: "Verify AI credentials",
    description:
      "Connect an organization-owned AI provider, verify the credential, and select the roles your agent requires.",
    href: "/docs/guides/ai/integrations",
    action: "AI integrations guide",
  },
  {
    number: "02",
    title: "Connect your SIP trunk",
    description:
      "Configure your existing carrier or PBX, add remote endpoints, secure ingress, and validate the trunk.",
    href: "/docs/guides/connectivity/sip-trunks",
    action: "SIP trunk guide",
  },
  {
    number: "03",
    title: "Create and check an agent",
    description:
      "Define instructions and an engine, bind its provider roles, then inspect readiness issues before accepting calls.",
    href: "/docs/guides/voice/voice-agents",
    action: "Voice Agent guide",
  },
  {
    number: "04",
    title: "Route an existing number",
    description:
      "Register a carrier-owned number, assign its trunk, and bind it to the agent for inbound calling.",
    href: "/docs/guides/voice/inbound-call",
    action: "Inbound call guide",
  },
];

export function DeveloperQuickstart() {
  return (
    <section
      id="quickstart"
      aria-labelledby="developer-quickstart-title"
      className="scroll-mt-24 border-b border-border"
    >
      <div className="mx-auto max-w-[1200px] px-5 py-24 sm:px-6 lg:px-8 lg:py-28">
        <div className="grid gap-8 lg:grid-cols-[1.05fr_0.95fr] lg:items-end">
          <div>
            <p className="font-mono text-xs uppercase tracking-[0.16em] text-brand-orange">
              Developer onboarding
            </p>
            <h2
              id="developer-quickstart-title"
              className="mt-5 max-w-2xl text-balance font-heading text-4xl font-semibold tracking-tight sm:text-5xl"
            >
              Four configurations before your first call.
            </h2>
          </div>
          <p className="max-w-lg leading-8 text-muted-foreground">
            Follow the same setup resources used by the Leamout Console.
            Configuration is not a successful call: provider readiness, SIP
            ingress, routing, and runtime capacity must all be in place.
          </p>
        </div>
        <ol className="mt-12 grid gap-px overflow-hidden rounded-xl border border-border bg-border md:grid-cols-2">
          {steps.map((step) => (
            <li key={step.number} className="flex flex-col bg-card p-7 sm:p-9">
              <span className="font-mono text-xs text-brand-orange">
                {step.number}
              </span>
              <h3 className="mt-8 font-heading text-xl font-semibold">
                {step.title}
              </h3>
              <p className="mt-3 flex-1 text-sm leading-7 text-muted-foreground">
                {step.description}
              </p>
              <Link
                href={step.href}
                className="mt-6 self-start text-sm font-medium text-foreground underline decoration-brand-orange underline-offset-4 hover:text-brand-orange"
              >
                {step.action}
              </Link>
            </li>
          ))}
        </ol>
        <div className="mt-6 grid gap-8 rounded-xl border border-border bg-card p-7 lg:grid-cols-[1fr_1.1fr] lg:items-center lg:p-9">
          <div>
            <p className="font-mono text-[11px] uppercase tracking-widest text-brand-orange">
              Verify before dialing
            </p>
            <h3 className="mt-3 font-heading text-2xl font-semibold">
              Check the agent readiness API.
            </h3>
            <p className="mt-3 text-sm leading-7 text-muted-foreground">
              An organization API token with permission and a saved agent ID
              are required. Check the `ready` value and address any reported
              issues before routing real traffic.
            </p>
            <Link
              href="/docs/quickstart"
              className="mt-5 inline-block text-sm font-medium text-foreground underline decoration-brand-orange underline-offset-4 hover:text-brand-orange"
            >
              Full API quickstart
            </Link>
          </div>
          <div className="min-w-0 rounded-lg border border-border bg-background p-5">
            <p className="mb-4 font-mono text-[11px] uppercase tracking-widest text-muted-foreground">
              GET /v1/voice-agents/:id/readiness
            </p>
            <pre className="overflow-x-auto font-mono text-xs leading-7 text-foreground">
              <code>{'curl -sS "https://api.leamout.com/v1/voice-agents/$VOICE_AGENT_ID/readiness" -H "Authorization: Bearer $LEAMOUT_TOKEN"'}</code>
            </pre>
            <p className="mt-4 text-xs leading-6 text-muted-foreground">
              This is a configuration check, not a live media or carrier test.
            </p>
          </div>
        </div>
      </div>
    </section>
  );
}
