const providerGroups = [
  {
    role: "Speech recognition",
    description: "Convert live speech into text for a composable agent.",
    providers: ["Deepgram", "AssemblyAI"],
  },
  {
    role: "Language models",
    description: "Generate responses and drive configured tool calls.",
    providers: ["Groq", "OpenAI"],
  },
  {
    role: "Speech synthesis",
    description: "Turn agent responses into streaming audio.",
    providers: ["Cartesia", "ElevenLabs"],
  },
  {
    role: "Realtime speech models",
    description: "Use one provider for a bidirectional speech session.",
    providers: ["OpenAI Realtime", "Gemini Live"],
  },
];

export function AIProviders() {
  return (
    <section className="mx-auto max-w-6xl px-5 py-20 sm:px-6 lg:px-8">
      <div className="max-w-3xl">
        <p className="font-mono text-xs uppercase tracking-widest text-brand-orange">
          AI providers
        </p>
        <h2 className="mt-4 text-balance text-3xl font-heading font-semibold tracking-tight sm:text-4xl">
          Choose the models that fit the conversation.
        </h2>
        <p className="mt-5 leading-8 text-muted-foreground">
          Leamout supports composable speech pipelines and integrated realtime
          models. Configure provider integrations using your own credentials and
          select them for each agent.
        </p>
      </div>
      <div className="mt-12 grid gap-4 sm:grid-cols-2">
        {providerGroups.map((group) => (
          <div
            key={group.role}
            className="rounded-xl border border-border bg-card p-6 sm:p-8"
          >
            <h3 className="text-xl font-semibold">{group.role}</h3>
            <p className="mt-3 text-sm leading-7 text-muted-foreground">
              {group.description}
            </p>
            <ul className="mt-7 flex flex-wrap gap-2">
              {group.providers.map((provider) => (
                <li
                  key={provider}
                  className="rounded-full border border-border bg-muted/30 px-4 py-2 text-sm font-medium"
                >
                  {provider}
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
      <p className="mt-6 text-sm leading-7 text-muted-foreground">
        These are adapters in the current Leamout provider catalog. Available
        models and features depend on the provider and your integration
        configuration.
      </p>
    </section>
  );
}
