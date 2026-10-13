import { DocumentLayout } from "@/components/legal/document-layout";

export function TermsOfService() {
  return (
    <DocumentLayout
      title="Terms of Service"
      description="Terms for accessing and using Leamout services."
    >
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
          software. Hosted-service terms and enterprise agreements are separate
          documents.
        </p>
      </section>
    </DocumentLayout>
  );
}
