import { WebhookDetails } from "@/components/webhooks/webhook-details";

export default async function Page({
  params,
}: PageProps<"/webhooks/[webhookId]">) {
  const { webhookId } = await params;

  return <WebhookDetails key={webhookId} webhookId={webhookId} />;
}
