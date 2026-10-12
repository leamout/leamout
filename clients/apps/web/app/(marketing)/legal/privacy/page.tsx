import { PageIntro } from "@/components/marketing/page-intro";
import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Privacy Policy",
  description:
    "Learn how Leamout collects, uses, and protects personal data, and understand your privacy rights.",
  noIndex: true,
  path: "/legal/privacy",
});

export default function PrivacyPolicyPage() {
  return (
    <div className="mx-auto max-w-6xl px-5 sm:px-6 lg:px-8">
      <PageIntro eyebrow="Trust and transparency" title="Privacy Policy">
        <p>Information about how Leamout handles personal data.</p>
      </PageIntro>
      <div className="mb-10 inline-flex rounded-full border border-border px-4 py-2 font-mono text-xs uppercase tracking-wider">
        Pending publication
      </div>
      <div className="max-w-3xl pb-20">
        <section className="border-t border-border py-8">
          <h2 className="text-xl font-semibold">Policy status</h2>
          <p className="mt-4 leading-8 text-muted-foreground">
            The privacy policy is being prepared. Details about the responsible
            legal entity, processing purposes, retention periods, international
            transfers, and rights requests will be published here.
          </p>
        </section>
        <section className="border-t border-border py-8">
          <h2 className="text-xl font-semibold">Deployment context</h2>
          <p className="mt-4 leading-8 text-muted-foreground">
            For self-hosted installations, the organization operating the
            deployment determines how its call and account data is handled.
            Consult that operator’s privacy information.
          </p>
        </section>
      </div>
    </div>
  );
}
