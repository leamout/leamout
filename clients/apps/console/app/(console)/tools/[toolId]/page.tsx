import { ToolDetails } from "@/components/tools/tool-details";

export default async function Page({
  params,
  searchParams,
}: PageProps<"/tools/[toolId]">) {
  const { toolId } = await params;
  const query = await searchParams;
  const agentId = typeof query.agent_id === "string" ? query.agent_id : "";

  return <ToolDetails key={toolId} toolId={toolId} agentId={agentId} />;
}
