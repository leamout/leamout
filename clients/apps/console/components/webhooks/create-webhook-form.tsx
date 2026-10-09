import Link from "next/link";
import { WebhookForm } from "@/components/webhooks/webhook-form";

export function CreateWebhookForm() {
  return (
    <div className="space-y-6">
      <Link
        href="/webhooks"
        className="text-sm text-muted-foreground hover:underline"
      >
        Back to webhooks
      </Link>
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">
          Create webhook endpoint
        </h1>
        <p className="text-muted-foreground">
          Choose where platform events are delivered and which events to
          receive.
        </p>
      </div>
      <WebhookForm creating />
    </div>
  );
}
