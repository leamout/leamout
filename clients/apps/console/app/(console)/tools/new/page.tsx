import { CreateToolForm } from "@/components/tools/create-tool-form";

export default async function Page({
  searchParams,
}: {
  searchParams: Promise<{ agent_id?: string | string[] }>;
}) {
  const params = await searchParams;
  const agentId = typeof params.agent_id === "string" ? params.agent_id : "";

  return <CreateToolForm agentId={agentId} />;
}
