import { DocumentLayout } from "@/components/legal/document-layout";

export function CookiePolicy() {
  return (
    <DocumentLayout
      title="Cookie Policy"
      description="Information about cookies and related website technologies."
    >
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
          authentication transaction metadata. Passwords and verification codes
          are not persisted in sessionStorage.
        </p>
      </section>
      <section className="border-t border-border py-8">
        <h2 className="text-xl font-semibold">Website measurement</h2>
        <p className="mt-4 leading-8 text-muted-foreground">
          The website includes Vercel Analytics and Speed Insights integrations.
          Their deployed behavior and any applicable consent requirements need
          to be confirmed before the policy is finalized.
        </p>
      </section>
    </DocumentLayout>
  );
}
