export function Mission() {
  return (
    <section
      className="grid gap-10 border-b border-border py-20 md:grid-cols-[0.8fr_1.2fr] md:gap-20"
      aria-labelledby="about-mission"
    >
      <div>
        <p className="font-mono text-xs uppercase tracking-[0.16em] text-brand-orange">
          Our mission
        </p>
        <h2
          id="about-mission"
          className="mt-5 max-w-md text-balance font-heading text-3xl font-semibold leading-tight tracking-tight sm:text-4xl"
        >
          Make real-time voice infrastructure something developers can build on.
        </h2>
      </div>
      <div className="max-w-2xl space-y-6 text-base leading-8 text-muted-foreground sm:text-lg">
        <p>
          Building an agent that can answer a call means handling much more than
          a prompt. It means connecting telephony to AI, managing a live
          session, and keeping the application in control as the conversation
          unfolds.
        </p>
        <p>
          Leamout brings those responsibilities together: SIP connectivity,
          media handling, agent orchestration, tool execution, and call
          operations. Developers focus on what their agents should do.
        </p>
        <p className="font-medium text-foreground">
          One service to run the conversation. Your choice of carrier and AI.
        </p>
      </div>
    </section>
  );
}
