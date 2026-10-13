import { DocumentLayout } from "@/components/legal/document-layout";

export function Subprocessors() {
  return (
    <DocumentLayout
      title="Subprocessors"
      description="Third-party processing information for Leamout services."
    >
      <section className="border-t border-border py-8">
        <h2 className="text-xl font-semibold">Register status</h2>
        <p className="mt-4 leading-8 text-muted-foreground">
          A verified subprocessor register has not yet been published. Provider
          identities, processing purposes, locations, and applicable notices
          will be listed here once confirmed.
        </p>
      </section>
      <section className="border-t border-border py-8">
        <h2 className="text-xl font-semibold">Providers you connect</h2>
        <p className="mt-4 leading-8 text-muted-foreground">
          Carrier and AI providers connected through your own accounts are
          selected by your organization. Their relationship to your data depends
          on your agreements and deployment.
        </p>
      </section>
    </DocumentLayout>
  );
}
