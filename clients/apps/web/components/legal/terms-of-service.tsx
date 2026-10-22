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
        <h2 className="text-xl font-semibold">Hosted service</h2>
        <p className="mt-4 leading-8 text-muted-foreground">
          Leamout is offered as a managed service. Commercial service terms,
          billing conditions and data processing requirements are being
          prepared; repository source licenses do not replace those terms.
        </p>
      </section>
    </DocumentLayout>
  );
}
