import Link from "next/link";

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
    <section className="mx-auto max-w-6xl px-5 py-20 sm:px-6 lg:px-8 lg:py-28">
      <div className="max-w-3xl">
        <p className="font-mono text-xs uppercase tracking-widest text-brand-orange">
          AI integrations
        </p>
        <h2 className="mt-4 text-balance font-heading text-3xl font-semibold tracking-tight sm:text-4xl">
          Choose the engine. Bind the providers.
        </h2>
        <p className="mt-5 leading-8 text-muted-foreground">
          Bring your own provider credentials and verify each integration before
          assigning it to an agent.
        </p>
      </div>
      <div className="mt-10 grid gap-px overflow-hidden rounded-xl border border-border bg-border md:grid-cols-2">
        <div className="bg-card p-7 sm:p-9">
          <p className="font-mono text-xs uppercase tracking-widest text-brand-orange">
            Realtime
          </p>
          <h3 className="mt-4 text-xl font-semibold">One provider binding.</h3>
          <p className="mt-3 text-sm leading-7 text-muted-foreground">
            A realtime speech provider handles the voice session through the
            agent's realtime role.
          </p>
        </div>
        <div className="bg-card p-7 sm:p-9">
          <p className="font-mono text-xs uppercase tracking-widest text-brand-orange">
            Composable
          </p>
          <h3 className="mt-4 text-xl font-semibold">
            STT + LLM + TTS bindings.
          </h3>
          <p className="mt-3 text-sm leading-7 text-muted-foreground">
            Independently select recognition, language, and speech synthesis
            providers for the same agent.
          </p>
        </div>
      </div>
      <div className="mt-8 grid gap-4 sm:grid-cols-2">
        {providerGroups.map((group) => (
          <div
            key={group.role}
            className="rounded-xl border border-border bg-card p-6 sm:p-8"
          >
            <h3 className="text-lg font-semibold">{group.role}</h3>
            <p className="mt-2 text-sm leading-7 text-muted-foreground">
              {group.description}
            </p>
            <ul className="mt-6 flex flex-wrap gap-2">
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
      <div className="mt-7 flex flex-wrap items-center justify-between gap-4">
        <p className="max-w-2xl text-sm leading-7 text-muted-foreground">
          Providers shown are in the current catalog. Credentials, supported
          models, and capabilities vary by integration. A saved credential must
          be verified before it can be relied on for calls.
        </p>
        <Link
          href="/docs/guides/ai/integrations"
          className="shrink-0 text-sm font-medium text-foreground underline decoration-brand-orange underline-offset-4 hover:text-brand-orange"
        >
          AI integration guide
        </Link>
      </div>
    </section>
  );
}
