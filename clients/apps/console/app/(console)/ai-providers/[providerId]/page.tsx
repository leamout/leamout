import { ProviderDetails } from "@/components/ai-providers/provider-details";

export default async function Page({
  params,
}: PageProps<"/ai-providers/[providerId]">) {
  const { providerId } = await params;

  return <ProviderDetails key={providerId} providerId={providerId} />;
}
