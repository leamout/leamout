import { AgentForm } from "@/components/agents/agent-form";

export function CreateAgentForm() {
  return (
    <div className="space-y-8">
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">Create agent</h1>
        <p className="text-muted-foreground">
          Choose a conversation engine, define the agent instructions, and save
          its configuration. Provider assignments and readiness are configured
          separately.
        </p>
      </div>
      <AgentForm creating />
    </div>
  );
}
