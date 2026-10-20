const waveformBars = [
  12, 19, 10, 27, 39, 21, 15, 35, 48, 30, 16, 22, 43, 54, 29, 16, 42,
  57, 38, 18, 31, 46, 27, 15, 33, 51, 36, 20, 14, 28, 44, 55, 34, 21, 38,
  26, 13, 20, 31, 18,
].map((height, index) => ({
  x: index * 8 + 2,
  height,
}));

const sessionEvents = [
  { time: "00:00", label: "SIP session established" },
  { time: "00:02", label: "Agent connected to audio" },
  { time: "00:04", label: "Speech turn detected" },
];

export function SessionPreview() {
  return (
    <div
      role="group"
      aria-label="Illustrative Leamout voice agent session"
      className="overflow-hidden rounded-2xl border border-border bg-card shadow-[0_24px_90px_rgba(0,0,0,0.2)]"
    >
      <div className="flex items-center justify-between gap-3 border-b border-border px-5 py-4">
        <div className="flex items-center gap-2">
          <span className="size-2 rounded-full bg-brand-orange" />
          <span className="font-mono text-[11px] font-medium uppercase tracking-[0.14em] text-foreground">
            Voice session
          </span>
        </div>
        <span className="font-mono text-[10px] uppercase tracking-wider text-muted-foreground">
          Illustrative preview
        </span>
      </div>
      <div className="space-y-5 p-5 sm:p-7">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
              Signaling
            </p>
            <p className="mt-2 flex items-center gap-2 text-sm font-medium">
              <span className="size-1.5 rounded-full bg-emerald-400" />
              SIP connected
            </p>
          </div>
          <div>
            <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
              Runtime
            </p>
            <p className="mt-2 text-sm font-medium">Voice agent active</p>
          </div>
        </div>
        <div className="rounded-xl border border-border bg-background p-4">
          <div className="flex items-center justify-between gap-3">
            <p className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
              Realtime audio
            </p>
            <span className="font-mono text-[10px] text-[#FF8E6D]">
              AUDIO / RTP
            </span>
          </div>
          <svg
            aria-hidden="true"
            className="mt-3 w-full"
            viewBox="0 0 320 70"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <line
              x1="0"
              x2="320"
              y1="35"
              y2="35"
              stroke="#514841"
              strokeDasharray="3 6"
            />
            {waveformBars.map((bar) => (
              <rect
                key={bar.x}
                x={bar.x}
                y={(70 - bar.height) / 2}
                width="3"
                height={bar.height}
                rx="1.5"
                fill={bar.x > 226 ? "#F43E01" : "#998C85"}
              />
            ))}
          </svg>
          <div className="mt-2 flex items-center justify-between font-mono text-[10px] text-muted-foreground">
            <span>INPUT STREAM</span>
            <span>OUTPUT STREAM</span>
          </div>
        </div>
        <div className="rounded-xl border border-border bg-background p-4">
          <div className="flex items-center justify-between gap-2">
            <h3 className="text-sm font-medium">Session events</h3>
            <span className="font-mono text-[10px] text-muted-foreground">
              EXAMPLE
            </span>
          </div>
          <ol className="mt-4 space-y-3">
            {sessionEvents.map((event) => (
              <li key={event.time} className="flex items-center gap-3">
                <span className="font-mono text-[10px] text-muted-foreground">
                  {event.time}
                </span>
                <span className="size-1.5 shrink-0 rounded-full bg-brand-orange" />
                <span className="text-xs text-foreground">{event.label}</span>
              </li>
            ))}
          </ol>
        </div>
        <div className="grid grid-cols-3 divide-x divide-border border-t border-border pt-4 text-center">
          <div>
            <p className="font-mono text-[10px] uppercase text-muted-foreground">
              Phone
            </p>
            <p className="mt-2 text-xs font-medium">SIP trunk</p>
          </div>
          <div>
            <p className="font-mono text-[10px] uppercase text-muted-foreground">
              Agent
            </p>
            <p className="mt-2 text-xs font-medium">Leamout</p>
          </div>
          <div>
            <p className="font-mono text-[10px] uppercase text-muted-foreground">
              AI
            </p>
            <p className="mt-2 text-xs font-medium">Your provider</p>
          </div>
        </div>
      </div>
    </div>
  );
}
