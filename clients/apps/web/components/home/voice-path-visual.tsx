const runtimeResponsibilities = [
  "SIP signaling",
  "Realtime media",
  "Agent execution",
  "Call lifecycle",
];

function FlowArrow() {
  return (
    <div className="flex items-center justify-center py-1 md:py-0">
      <svg
        aria-hidden="true"
        viewBox="0 0 24 24"
        fill="none"
        className="size-5 text-brand-orange md:hidden"
      >
        <path
          d="M12 3v17m0 0-6-6m6 6 6-6"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinecap="square"
          strokeLinejoin="round"
        />
      </svg>
      <svg
        aria-hidden="true"
        viewBox="0 0 24 24"
        fill="none"
        className="hidden h-5 w-6 text-brand-orange md:block"
      >
        <path
          d="M2 12h19m0 0-6-6m6 6-6 6"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinecap="square"
          strokeLinejoin="round"
        />
      </svg>
    </div>
  );
}

export function VoicePathVisual() {
  return (
    <figure className="relative overflow-hidden border border-foreground/15 bg-white shadow-[0_28px_70px_rgba(16,19,22,0.08)]">
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border px-4 py-4 sm:px-6">
        <div className="flex items-center gap-3">
          <span
            aria-hidden="true"
            className="flex size-7 items-center justify-center border border-border font-mono text-[10px] text-brand-orange"
          >
            01
          </span>
          <figcaption className="font-mono text-[11px] font-medium uppercase tracking-[0.14em] text-foreground">
            The voice path
          </figcaption>
        </div>
        <span className="font-mono text-[9px] uppercase tracking-wider text-muted-foreground sm:text-[10px]">
          Conceptual system map
        </span>
      </div>

      <div className="p-4 sm:p-6">
        <div className="grid items-stretch gap-2 md:grid-cols-[minmax(0,1fr)_28px_minmax(0,1.15fr)_28px_minmax(0,1fr)] md:gap-1">
          <div className="flex min-h-[148px] flex-col border border-border bg-[#fbfbf9] p-4 sm:p-5">
            <p className="font-mono text-[10px] uppercase tracking-[0.12em] text-muted-foreground">
              01 / Telephony
            </p>
            <div className="mt-5 flex items-center gap-2">
              <svg
                aria-hidden="true"
                viewBox="0 0 24 24"
                fill="none"
                className="size-5 shrink-0 text-foreground"
              >
                <path
                  d="M5 4h4l2 5-2.5 1.5a15 15 0 0 0 5 5L15 13l5 2v4a1 1 0 0 1-1 1C10 20 4 14 4 5a1 1 0 0 1 1-1Z"
                  stroke="currentColor"
                  strokeWidth="1.4"
                  strokeLinejoin="round"
                />
              </svg>
              <h3 className="font-heading text-lg font-semibold tracking-tight">
                SIP carrier
              </h3>
            </div>
            <p className="mt-3 text-xs leading-5 text-muted-foreground">
              Your numbers, trunk, or PBX.
            </p>
          </div>

          <FlowArrow />

          <div className="relative flex min-h-[148px] flex-col border border-brand-orange/60 bg-[#fff5ef] p-4 sm:p-5">
            <span
              aria-hidden="true"
              className="absolute inset-x-0 top-0 h-1 bg-brand-orange"
            />
            <p className="font-mono text-[10px] uppercase tracking-[0.12em] text-brand-orange">
              02 / Managed runtime
            </p>
            <div className="mt-5 flex items-center gap-2">
              <span
                aria-hidden="true"
                className="flex size-5 items-center justify-center bg-brand-orange text-[10px] font-bold text-brand-graphite"
              >
                L
              </span>
              <h3 className="font-heading text-lg font-semibold tracking-tight">
                Leamout
              </h3>
            </div>
            <p className="mt-3 text-xs leading-5 text-muted-foreground">
              Operates the session between SIP and AI.
            </p>
          </div>

          <FlowArrow />

          <div className="flex min-h-[148px] flex-col border border-border bg-[#fbfbf9] p-4 sm:p-5">
            <p className="font-mono text-[10px] uppercase tracking-[0.12em] text-muted-foreground">
              03 / Intelligence
            </p>
            <div className="mt-5 flex items-center gap-2">
              <svg
                aria-hidden="true"
                viewBox="0 0 24 24"
                fill="none"
                className="size-5 shrink-0 text-foreground"
              >
                <path
                  d="M12 3.5 20 8v8l-8 4.5L4 16V8l8-4.5Z M4 8l8 4.5L20 8 M12 12.5v8"
                  stroke="currentColor"
                  strokeWidth="1.4"
                  strokeLinejoin="round"
                />
              </svg>
              <h3 className="font-heading text-lg font-semibold tracking-tight">
                AI providers
              </h3>
            </div>
            <p className="mt-3 text-xs leading-5 text-muted-foreground">
              Realtime model or STT, LLM, and TTS.
            </p>
          </div>
        </div>

        <div className="mt-6 border-t border-border pt-5">
          <div className="flex flex-wrap items-baseline justify-between gap-3">
            <p className="font-mono text-[10px] uppercase tracking-[0.14em] text-muted-foreground">
              What the runtime owns
            </p>
            <p className="text-[10px] text-muted-foreground">
              Illustrative, not live traffic
            </p>
          </div>
          <ul className="mt-4 grid grid-cols-2 gap-x-4 gap-y-3 sm:grid-cols-4">
            {runtimeResponsibilities.map((item) => (
              <li key={item} className="flex items-start gap-2 text-xs">
                <span
                  aria-hidden="true"
                  className="mt-0.5 h-3 w-0.5 shrink-0 bg-brand-orange"
                />
                <span className="leading-5">{item}</span>
              </li>
            ))}
          </ul>
        </div>
      </div>

      <div className="flex flex-wrap items-center justify-between gap-2 border-t border-border bg-[#fbfbf9] px-4 py-3 font-mono text-[9px] uppercase tracking-wider text-muted-foreground sm:px-6">
        <span>Keep your carrier and provider relationships</span>
        <span>Call path / architecture view</span>
      </div>
    </figure>
  );
}
