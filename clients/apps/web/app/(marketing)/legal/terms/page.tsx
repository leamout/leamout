import { PageIntro } from "@/components/marketing/page-intro";
import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Terms of Service",
  description:
    "Read the terms governing access to and use of Leamout services.",
  noIndex: true,
  path: "/legal/terms",
});

export default function TermsofServicePage() {
  return (
    <div className="mx-auto max-w-6xl px-5 sm:px-6 lg:px-8">
      <PageIntro eyebrow="Trust and transparency" title="Terms of Service">
        <p>Terms for accessing and using Leamout services.</p>
      </PageIntro>
      <div className="mb-10 inline-flex rounded-full border border-border px-4 py-2 font-mono text-xs uppercase tracking-wider">
        Pending publication
      </div>
      <div className="max-w-3xl pb-20">
        <section className="border-t border-border py-8">
          <h2 className="text-xl font-semibold">Publication status</h2>
          <p className="mt-4 leading-8 text-muted-foreground">
            Service terms are being prepared. This page is not an agreement or a
            statement of subscription, refund, warranty, or service-level
            commitments.
          </p>
        </section>
        <section className="border-t border-border py-8">
          <h2 className="text-xl font-semibold">Open-source software</h2>
          <p className="mt-4 leading-8 text-muted-foreground">
            The repository contains the license governing the open-source
            software. Hosted-service terms and enterprise agreements are
            separate documents.
          </p>
        </section>
      </div>
    </div>
  );
}
