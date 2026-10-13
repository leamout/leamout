import { Button } from "@leamout/ui/components/button";
import Link from "next/link";

export function Enterprise() {
  return (
    <section className="my-6 flex flex-col justify-between gap-6 rounded-xl bg-muted/40 p-8 sm:flex-row sm:items-center">
      <div>
        <h2 className="text-2xl font-semibold">Enterprise</h2>
        <p className="mt-2 max-w-2xl leading-7 text-muted-foreground">
          Custom capacity, SSO, SCIM, private networking, and deployment
          requirements. Scope and pricing are agreed for your organization.
        </p>
      </div>
      <Button
        variant="outline"
        nativeButton={false}
        render={<Link href="/contact" />}
      >
        Contact us
      </Button>
    </section>
  );
}
