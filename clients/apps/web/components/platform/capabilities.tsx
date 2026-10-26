const capabilities = [
  {
    title: "Programmable call control",
    description:
      "Answer, transfer, hold, record, and manage calls while preserving the connection to the voice agent.",
  },
  {
    title: "Realtime AI orchestration",
    description:
      "Run speech-to-speech sessions or a composed speech recognition, language model, and synthesis pipeline.",
  },
  {
    title: "Tools and application events",
    description:
      "Let agents execute configured tools and use webhooks to connect call activity to your applications.",
  },
  {
    title: "Operational visibility",
    description:
      "Inspect persisted call milestones, SIP and routing identifiers, and linked recordings where available.",
  },
];

export function PlatformCapabilities() {
  return (
    <section className="mx-auto max-w-6xl px-5 py-20 sm:px-6 lg:px-8">
      <div className="grid gap-12 lg:grid-cols-[0.85fr_1.15fr]">
        <div>
          <p className="font-mono text-xs uppercase tracking-widest text-brand-orange">
            Built for live conversations
          </p>
          <h2 className="mt-4 text-balance text-3xl font-heading font-semibold tracking-tight sm:text-4xl">
            The operational layer between telephony and AI.
          </h2>
          <p className="mt-6 leading-8 text-muted-foreground">
            Leamout operates the managed runtime. Your carrier continues to
            provide telephony, and your AI providers supply model execution.
            Your team controls agent behavior and integrations.
          </p>
        </div>
        <div className="grid gap-x-8 sm:grid-cols-2">
          {capabilities.map((capability) => (
            <div key={capability.title} className="border-t border-border py-7">
              <h3 className="text-lg font-semibold">{capability.title}</h3>
              <p className="mt-3 leading-7 text-muted-foreground">
                {capability.description}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
