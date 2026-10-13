import { Suspense } from "react";
import { BillingSettings } from "@/components/billing/billing-settings";

export default function Page() {
  return (
    <Suspense
      fallback={
        <p role="status" className="text-sm text-muted-foreground">
          Loading billing settings…
        </p>
      }
    >
      <BillingSettings />
    </Suspense>
  );
}
