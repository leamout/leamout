import { AgentEditor } from "@/components/resources/agent-editor";
export default async function Page({params}:PageProps<"/agents/[agentId]">){const {agentId}=await params;return <AgentEditor id={agentId}/>;}
