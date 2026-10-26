import Link from "next/link";

const callTabs = ["Lifecycle", "Diagnostics", "Recordings", "Conversation"];

const diagnosticFields = [
  "SIP Call-ID",
  "Voice Agent",
  "SIP trunk",
  "Trunk endpoint ID",
  "Media state",
  "Hangup reason",
];

export function ControlPlane() {
  return (
    <section id="console" className="scroll-mt-24 border-b border-border bg-card">
      <div className="mx-auto grid max-w-[1200px] gap-14 px-5 py-24 sm:px-6 lg:grid-cols-[1.1fr_0.9fr] lg:items-center lg:gap-20 lg:px-8 lg:py-28">
        <section
          aria-labelledby="call-detail-preview-title"
          className="min-w-0 overflow-hidden rounded-2xl border border-border bg-background"
        >
          <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border px-5 py-4">
            <h3
              id="call-detail-preview-title"
              className="font-mono text-[11px] uppercase tracking-widest"
            >
              Call details
            </h3>
            <span className="font-mono text-[10px] text-muted-foreground">
              Console interface reference
            </span>
          </div>
          <div className="px-5 pt-5 sm:px-6">
            <p className="text-xs leading-6 text-muted-foreground">
              Inspect the persisted details of a call in your organization.
              No real call data is shown here.
            </p>
            <div className="mt-5 grid grid-cols-2 gap-2 sm:grid-cols-4">
              {["State", "Media state", "Direction", "Talk time"].map(
                (field) => (
                  <div
                    key={field}
                    className="rounded-md border border-border bg-card px-3 py-3"
                  >
                    <p className="font-mono text-[10px] text-muted-foreground">
                      {field}
                    </p>
                    <p className="mt-2 text-xs text-foreground">Call record</p>
                  </div>
                ),
              )}
            </div>
          </div>
          <div className="mx-5 mt-6 flex flex-wrap gap-x-5 gap-y-2 border-b border-border text-xs sm:mx-6">
            {callTabs.map((tab) => (
              <span
                key={tab}
                className={
                  tab === "Diagnostics"
                    ? "border-b-2 border-brand-orange pb-3 font-medium text-foreground"
                    : "pb-3 text-muted-foreground"
                }
              >
                {tab}
              </span>
            ))}
          </div>
          <div className="p-5 sm:p-6">
            <h4 className="text-sm font-semibold">
              SIP and routing identifiers
            </h4>
            <p className="mt-2 text-xs leading-6 text-muted-foreground">
              Fields available in the call diagnostics panel.
            </p>
            <ul className="mt-4 grid gap-2 sm:grid-cols-2">
              {diagnosticFields.map((field) => (
                <li
                  key={field}
                  className="border-l-2 border-brand-orange bg-card px-3 py-2 text-xs text-foreground"
                >
                  {field}
                </li>
              ))}
            </ul>
            <p className="mt-5 border-t border-border pt-4 text-xs leading-6 text-muted-foreground">
              The current call API does not expose SIP message traces,
              per-turn AI latency, transcripts, or tool execution traces.
            </p>
          </div>
        </section>
        <div>
          <p className="font-mono text-xs uppercase tracking-[0.16em] text-brand-orange">
            Call operations
          </p>
          <h2 className="mt-5 text-balance font-heading text-4xl font-semibold tracking-tight sm:text-5xl">
            Debug the call, not just the prompt.
          </h2>
          <p className="mt-6 leading-8 text-muted-foreground">
            The Console exposes persisted call lifecycle timestamps, SIP and
            routing identifiers, call state, and linked recordings where
            available. Inspect real call records when your organization has
            traffic.
          </p>
          <ul className="mt-8 space-y-3 text-sm text-foreground">
            <li className="border-l-2 border-brand-orange pl-4">
              Call lifecycle and persisted milestones
            </li>
            <li className="border-l-2 border-brand-orange pl-4">
              SIP Call-ID, trunk attribution, and media state
            </li>
            <li className="border-l-2 border-brand-orange pl-4">
              Recording links for calls with recordings
            </li>
          </ul>
          <Link
            href="/docs/guides/voice/call-events"
            className="mt-8 inline-block text-sm font-medium text-foreground underline decoration-brand-orange underline-offset-4 hover:text-brand-orange"
          >
            Understand call lifecycle events
          </Link>
        </div>
      </div>
    </section>
  );
}
