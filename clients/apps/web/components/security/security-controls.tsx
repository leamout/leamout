const controls = [
  {
    number: "01",
    title: "Organization boundaries",
    description:
      "Resources, provider integrations, and access checks are tied to the authenticated organization context. Authorization is enforced on the server, not just by the console.",
  },
  {
    number: "02",
    title: "Secret handling",
    description:
      "Organization-owned provider credentials are encrypted at rest and excluded from public API responses. Secrets should not be copied into session snapshots, logs, or event payloads.",
  },
  {
    number: "03",
    title: "Agent and webhook execution",
    description:
      "Tool execution is scoped to configured permissions, and outbound webhook destinations are subject to egress policy. Voice agents must not receive unrestricted infrastructure access.",
  },
  {
    number: "04",
    title: "Call data and retention",
    description:
      "Call and recording access is organization-scoped. Retention follows the subscribed plan, with configurable retention policies on eligible plans.",
  },
];

export function SecurityControls() {
  return (
    <section
      className="border-b border-border py-20"
      aria-labelledby="security-controls-title"
    >
      <div className="mb-12 grid gap-6 md:grid-cols-2 md:items-end">
        <div>
          <p className="font-mono text-xs uppercase tracking-[0.16em] text-brand-orange">
            The security model
          </p>
          <h2
            id="security-controls-title"
            className="mt-5 text-balance font-heading text-3xl font-semibold tracking-tight sm:text-4xl"
          >
            Built around explicit boundaries.
          </h2>
        </div>
        <p className="max-w-md leading-7 text-muted-foreground">
          Security is a system property, from the control plane to the media
          runtime and the external services an agent can reach.
        </p>
      </div>
      <div className="border-t border-border">
        {controls.map(({ number, title, description }) => (
          <div
            key={number}
            className="grid gap-4 border-b border-border py-7 md:grid-cols-[3rem_0.8fr_1.2fr] md:gap-12"
          >
            <span className="font-mono text-xs text-brand-orange">
              {number}
            </span>
            <h3 className="font-heading text-xl font-medium tracking-tight sm:text-2xl">
              {title}
            </h3>
            <p className="max-w-xl leading-7 text-muted-foreground">
              {description}
            </p>
          </div>
        ))}
      </div>
    </section>
  );
}
