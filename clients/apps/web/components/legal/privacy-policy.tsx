import { DocumentLayout } from "@/components/legal/document-layout";

export function PrivacyPolicy() {
  return (
    <DocumentLayout
      title="Privacy Policy"
      description="Information about how Leamout handles personal data."
    >
      <section className="border-t border-border py-8">
        <h2 className="text-xl font-semibold">Policy status</h2>
        <p className="mt-4 leading-8 text-muted-foreground">
          The privacy policy is being prepared. Details about the responsible
          legal entity, processing purposes, retention periods, international
          transfers, and rights requests will be published here.
        </p>
      </section>
      <section className="border-t border-border py-8">
        <h2 className="text-xl font-semibold">Managed service context</h2>
        <p className="mt-4 leading-8 text-muted-foreground">
          Leamout is operated by Leamout. The approved privacy policy
          will describe how service data is processed, protected and retained,
          including the responsibilities of customers and service providers.
        </p>
      </section>
    </DocumentLayout>
  );
}
