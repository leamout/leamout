import { PageIntro } from "@/components/marketing/page-intro";
import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Subprocessors",
  description:
    "Review the third-party service providers that process personal data on behalf of Leamout.",
  noIndex: true,
  path: "/subprocessors",
});

export default function SubprocessorsPage() {
  return (
    <div className="mx-auto max-w-6xl px-5 sm:px-6 lg:px-8">
      <PageIntro eyebrow="Trust and transparency" title="Subprocessors">
        <p>Third-party processing information for Leamout services.</p>
      </PageIntro>
      <div className="mb-10 inline-flex rounded-full border border-border px-4 py-2 font-mono text-xs uppercase tracking-wider">
        Pending publication
      </div>
      <div className="max-w-3xl pb-20">
        <section className="border-t border-border py-8">
          <h2 className="text-xl font-semibold">Register status</h2>
          <p className="mt-4 leading-8 text-muted-foreground">
            A verified subprocessor register has not yet been published.
            Provider identities, processing purposes, locations, and applicable
            notices will be listed here once confirmed.
          </p>
        </section>
        <section className="border-t border-border py-8">
          <h2 className="text-xl font-semibold">Providers you connect</h2>
          <p className="mt-4 leading-8 text-muted-foreground">
            Carrier and AI providers connected through your own accounts are
            selected by your organization. Their relationship to your data
            depends on your agreements and deployment.
          </p>
        </section>
      </div>
    </div>
  );
}
