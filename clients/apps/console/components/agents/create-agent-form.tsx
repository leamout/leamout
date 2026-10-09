import { AgentForm } from "@/components/agents/agent-form";

export function CreateAgentForm() {
  return (
    <div className="space-y-8">
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">Create agent</h1>
        <p className="text-muted-foreground">
          Start with a name and description. Models, voice, and call behavior
          are configured in agent details.
        </p>
      </div>
      <AgentForm creating />
    </div>
  );
}
