export function Principles() {
  return (
    <div className="grid gap-6 py-8 md:grid-cols-3">
      {[
        [
          "Bring your own",
          "Connect your SIP carrier or PBX and the AI providers that fit your application.",
        ],
        [
          "Build programmatically",
          "Use APIs, tools, and webhooks to connect voice agents to the rest of your product.",
        ],
        [
          "Choose your deployment",
          "Use the same platform architecture for self-hosted and cloud deployments.",
        ],
      ].map(([title, text]) => (
        <section key={title} className="rounded-xl border border-border p-6">
          <h2 className="text-xl font-semibold">{title}</h2>
          <p className="mt-4 leading-7 text-muted-foreground">{text}</p>
        </section>
      ))}
    </div>
  );
}
