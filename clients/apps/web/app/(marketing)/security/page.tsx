import Link from "next/link";
import { PageIntro } from "@/components/marketing/page-intro";
import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Security",
  description:
    "Learn about security in Leamout, including access controls, credential handling, and deployment considerations for autonomous voice agents.",
  path: "/security",
});

export default function SecurityPage() {
  return (
    <div className="mx-auto max-w-6xl px-5 sm:px-6 lg:px-8">
      <PageIntro
        eyebrow="Security"
        title="Control starts with clear boundaries."
      >
        <p>
          Leamout separates organizations, scopes provider access, and gives
          operators control over how their voice infrastructure is deployed.
        </p>
      </PageIntro>
      <div className="grid gap-x-12 md:grid-cols-2">
        {[
          [
            "Organization-scoped access",
            "Resources and provider integrations belong to organizations. API access is checked against the authenticated identity and organization context.",
          ],
          [
            "Provider credentials",
            "Provider secrets are encrypted at rest and are not returned through the public API. Operators are responsible for protecting deployment encryption keys.",
          ],
          [
            "Retention controls",
            "Retention limits follow your plan. Configurable policies are available on eligible plans, with scheduled cleanup for supported resources.",
          ],
          [
            "Deployment responsibility",
            "Self-hosted operators manage their network, certificates, updates, backups, and infrastructure access. Deployment configuration is part of the security boundary.",
          ],
        ].map(([title, text]) => (
          <section key={title} className="border-t border-border py-8">
            <h2 className="text-xl font-semibold">{title}</h2>
            <p className="mt-4 leading-8 text-muted-foreground">{text}</p>
          </section>
        ))}
      </div>
      <section className="my-16 rounded-xl bg-muted/40 p-8">
        <h2 className="text-2xl font-semibold">Review the implementation.</h2>
        <p className="my-4 max-w-2xl leading-7 text-muted-foreground">
          Evaluate the platform against your own requirements. This overview
          does not assert a compliance certification or contractual security
          guarantee.
        </p>
        <Link
          href="/docs"
          className="font-medium underline decoration-brand-orange underline-offset-8"
        >
          Explore the documentation
        </Link>
      </section>
    </div>
  );
}
