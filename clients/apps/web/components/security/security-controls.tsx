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
          "Provider secrets are encrypted at rest and are not returned through the public API. Leamout operators protect the platform encryption keys.",
        ],
        [
          "Retention controls",
          "Retention limits follow your plan. Configurable policies are available on eligible plans, with scheduled cleanup for supported resources.",
        ],
        [
          "Managed infrastructure",
          "Leamout operates and maintains the hosting infrastructure, service certificates, updates and platform backups. Customers protect access to their own carrier and AI provider accounts.",
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
