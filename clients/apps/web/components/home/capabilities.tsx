export function Capabilities() {
  return (
    <section className="mx-auto max-w-6xl px-5 py-20 sm:px-6 lg:px-8">
      <h2 className="text-3xl font-semibold tracking-tight sm:text-4xl">
        The conversation is only the beginning.
      </h2>
      <div className="mt-10 grid gap-x-12 md:grid-cols-2">
        {[
          [
            "Call control",
            "Route calls, send DTMF, manage playback, and hand conversations off to people.",
          ],
          [
            "AI orchestration",
            "Compose speech recognition, language models, and synthesis, or connect an integrated realtime provider.",
          ],
          [
            "Tools and events",
            "Let agents act through configured tools. Connect application workflows through event-driven webhooks.",
          ],
          [
            "Operational visibility",
            "Review calls, recordings, and conversation history alongside the configuration that shaped each session.",
          ],
        ].map(([title, text]) => (
          <div key={title} className="border-t border-border py-7">
            <h3 className="text-xl font-semibold">{title}</h3>
            <p className="mt-3 leading-7 text-muted-foreground">{text}</p>
          </div>
        ))}
      </div>
    </section>
  );
}
