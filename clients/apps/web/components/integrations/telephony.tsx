export function IntegrationsTelephony() {
  return (
    <section className="theme-light border-y border-border bg-background text-foreground">
      <div className="mx-auto grid max-w-6xl gap-10 px-5 py-20 sm:px-6 lg:grid-cols-[1fr_1fr] lg:gap-16 lg:px-8">
        <div>
          <p className="font-mono text-xs uppercase tracking-widest text-foreground">
            Telephony
          </p>
          <h2 className="mt-4 text-balance text-3xl font-semibold tracking-tight sm:text-4xl">
            Keep your carrier. Connect over SIP.
          </h2>
          <p className="mt-6 leading-8 text-muted-foreground">
            Use your existing carrier account, phone numbers, and SIP trunks.
            Leamout provides the voice agent runtime and programmable call
            control around that connection.
          </p>
          <p className="mt-4 leading-8 text-muted-foreground">
            Carrier charges and telephone number management remain with your
            provider.
          </p>
        </div>
        <div className="self-center rounded-xl border border-border bg-card p-6 sm:p-8">
          <p className="font-mono text-xs uppercase tracking-widest text-muted-foreground">
            Connection path
          </p>
          <ol className="mt-6 space-y-0">
            {[
              ["01", "Your carrier or PBX", "Existing numbers and SIP service"],
              ["02", "Leamout SIP edge", "Call signaling and realtime media"],
              ["03", "Leamout voice agent", "Conversation, tools, and call control"],
            ].map(([number, title, description], index) => (
              <li key={number} className="relative flex gap-5 pb-7 last:pb-0">
                {index < 2 && (
                  <span
                    aria-hidden="true"
                    className="absolute left-[15px] top-9 h-[calc(100%-2.25rem)] w-px bg-border"
                  />
                )}
                <span className="relative z-10 flex size-8 shrink-0 items-center justify-center rounded-full border border-brand-orange bg-background font-mono text-xs text-foreground">
                  {number}
                </span>
                <div>
                  <h3 className="font-semibold">{title}</h3>
                  <p className="mt-1 text-sm leading-6 text-muted-foreground">
                    {description}
                  </p>
                </div>
              </li>
            ))}
          </ol>
        </div>
      </div>
    </section>
  );
}
