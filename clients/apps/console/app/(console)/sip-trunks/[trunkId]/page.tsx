import { TrunkDetails } from "@/components/sip-trunks/trunk-details";

export default async function Page({
  params,
}: PageProps<"/sip-trunks/[trunkId]">) {
  const { trunkId } = await params;

  return <TrunkDetails trunkId={trunkId} />;
}
