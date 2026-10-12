import { MarketingActions } from "@/components/marketing/actions";
import { PageIntro } from "@/components/marketing/page-intro";
import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "About",
  description:
    "Learn about Leamout and our mission to help developers and enterprises run autonomous voice agents with their own carriers and AI providers.",
  path: "/about",
});

export default function AboutPage() {
  return (
    <div className="mx-auto max-w-6xl px-5 sm:px-6 lg:px-8">
      <PageIntro
        eyebrow="About Leamout"
        title="Infrastructure for the conversations ahead."
      >
        <p>
          We are building the runtime and control plane that developers and
          enterprises use to run their own autonomous voice agents.
        </p>
      </PageIntro>
      <div className="grid gap-12 border-t border-border py-12 md:grid-cols-2">
        <h2 className="text-3xl font-semibold tracking-tight">
          A platform built around ownership.
        </h2>
        <div className="space-y-5 leading-8 text-muted-foreground">
          <p>
            Voice agents connect two worlds: telephony infrastructure and AI
            models. Leamout brings them together while keeping your carrier
            accounts, phone numbers, and AI provider relationships in your
            hands.
          </p>
          <p>
            Our focus is the operational layer: configuring agents, running
            conversations, executing tools, and understanding what happened
            during a call.
          </p>
        </div>
      </div>
      <div className="grid gap-6 py-8 md:grid-cols-3">
        {[
          [
            "Bring your own",
            "Connect your SIP carrier or PBX and the AI providers that fit your application.",
          ],
          [
            "Build programmatically",
            "Use APIs, tools, and webhooks to connect voice agents to the rest of your product.",
          ],
          [
            "Choose your deployment",
            "Use the same platform architecture for self-hosted and cloud deployments.",
          ],
        ].map(([title, text]) => (
          <section key={title} className="rounded-xl border border-border p-6">
            <h2 className="text-xl font-semibold">{title}</h2>
            <p className="mt-4 leading-7 text-muted-foreground">{text}</p>
          </section>
        ))}
      </div>
      <div className="py-16">
        <MarketingActions />
      </div>
    </div>
  );
}
