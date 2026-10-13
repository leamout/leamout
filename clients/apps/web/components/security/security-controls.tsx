export function SecurityControls() {
  return (
    <div className="grid gap-x-12 md:grid-cols-2">
      {[
        [
          "Organization-scoped access",
          "Resources and provider integrations belong to organizations. API access is checked against the authenticated identity and organization context.",
        ],
        [
          "Provider credentials",
          "Provider secrets are encrypted at rest and are not returned through the public API. Operators are responsible for protecting deployment encryption keys.",
        ],
        [
          "Retention controls",
          "Retention limits follow your plan. Configurable policies are available on eligible plans, with scheduled cleanup for supported resources.",
        ],
        [
          "Deployment responsibility",
          "Self-hosted operators manage their network, certificates, updates, backups, and infrastructure access. Deployment configuration is part of the security boundary.",
        ],
      ].map(([title, text]) => (
        <section key={title} className="border-t border-border py-8">
          <h2 className="text-xl font-semibold">{title}</h2>
          <p className="mt-4 leading-8 text-muted-foreground">{text}</p>
        </section>
      ))}
    </div>
  );
}
