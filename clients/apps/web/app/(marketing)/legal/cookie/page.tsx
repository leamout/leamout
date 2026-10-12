import { PageIntro } from "@/components/marketing/page-intro";
import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Cookie Policy",
  description:
    "Learn how Leamout uses cookies and similar technologies and how you can manage your preferences.",
  noIndex: true,
  path: "/legal/cookie",
});

export default function CookiePolicyPage() {
  return (
    <div className="mx-auto max-w-6xl px-5 sm:px-6 lg:px-8">
      <PageIntro eyebrow="Trust and transparency" title="Cookie Policy">
        <p>Information about cookies and related website technologies.</p>
      </PageIntro>
      <div className="mb-10 inline-flex rounded-full border border-border px-4 py-2 font-mono text-xs uppercase tracking-wider">
        Pending publication
      </div>
      <div className="max-w-3xl pb-20">
        <section className="border-t border-border py-8">
          <h2 className="text-xl font-semibold">Policy status</h2>
          <p className="mt-4 leading-8 text-muted-foreground">
            The cookie policy is being prepared. A verified inventory of browser
            storage and applicable preference controls will be published here.
          </p>
        </section>
        <section className="border-t border-border py-8">
          <h2 className="text-xl font-semibold">Console authentication</h2>
          <p className="mt-4 leading-8 text-muted-foreground">
            The console uses a session cookie for sign-in and sessionStorage for
            authentication transaction metadata. Passwords and verification
            codes are not persisted in sessionStorage.
          </p>
        </section>
        <section className="border-t border-border py-8">
          <h2 className="text-xl font-semibold">Website measurement</h2>
          <p className="mt-4 leading-8 text-muted-foreground">
            The website includes Vercel Analytics and Speed Insights
            integrations. Their deployed behavior and any applicable consent
            requirements need to be confirmed before the policy is finalized.
          </p>
        </section>
      </div>
    </div>
  );
}
