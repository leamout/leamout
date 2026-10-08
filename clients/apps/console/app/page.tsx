import Link from "next/link";

const steps = [
  {
    title: "Connect an AI provider",
    description:
      "Bring the credentials for your speech and language providers.",
    href: "/connections",
  },
  {
    title: "Connect a SIP trunk",
    description: "Connect your carrier to handle inbound and outbound calls.",
    href: "/connections",
  },
  {
    title: "Create an agent",
    description: "Choose its voice, model, instructions, and tools.",
    href: "/agents",
  },
  {
    title: "Inspect your calls",
    description: "Review call outcomes and conversation activity.",
    href: "/calls",
  },
];

export default function Overview() {
  return (
    <div className="space-y-8">
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">Overview</h1>
        <p className="text-muted-foreground">
          Set up your voice runtime, one connection at a time.
        </p>
      </div>
      <div className="grid gap-4 md:grid-cols-2">
        {steps.map((step) => (
          <Link
            key={step.title}
            href={step.href}
            className="rounded-xl border bg-card p-6 transition-colors hover:bg-accent focus-visible:outline-2"
          >
            <h2 className="font-medium">{step.title}</h2>
            <p className="mt-2 text-sm text-muted-foreground">
              {step.description}
            </p>
          </Link>
        ))}
      </div>
      <p className="text-sm text-muted-foreground">
        This preview establishes navigation. Live organization data will follow.
      </p>
    </div>
  );
}
