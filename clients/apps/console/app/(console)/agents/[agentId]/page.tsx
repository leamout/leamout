import { AgentDetails } from "@/components/agents/agent-details";

export default async function Page({ params }: PageProps<"/agents/[agentId]">) {
  const { agentId } = await params;

  return <AgentDetails agentId={agentId} />;
}
