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
        <h2 className="text-xl font-semibold">Deployment context</h2>
        <p className="mt-4 leading-8 text-muted-foreground">
          For self-hosted installations, the organization operating the
          deployment determines how its call and account data is handled.
          Consult that operator’s privacy information.
        </p>
      </section>
    </DocumentLayout>
  );
}
