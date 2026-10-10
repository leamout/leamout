import Link from "next/link";
import type { Tool } from "@/components/tools/types";

export function AgentTools({
  agentId,
  tools = [],
}: {
  agentId: string;
  tools?: Tool[];
}) {
  const owned = tools.filter((tool) => tool.agent.id === agentId);
  return (
    <section className="space-y-6" aria-labelledby="agent-tools-title">
      <div className="space-y-2">
        <h2 id="agent-tools-title" className="text-lg font-semibold">
          Agent tools
        </h2>
        <p className="text-sm text-muted-foreground">
          Tools are owned by an agent. Review their configuration and
          availability during conversations.
        </p>
        <Link href="/tools/new" className="text-sm hover:underline">
          Create a tool
        </Link>
      </div>
      {owned.length ? (
        <ul className="space-y-3">
          {owned.map((tool) => (
            <li key={tool.id} className="space-y-2 rounded-lg border p-4">
              <Link
                href={`/tools/${encodeURIComponent(tool.id)}`}
                className="font-medium hover:underline"
              >
                {tool.name}
              </Link>
              <p className="text-sm text-muted-foreground">
                {tool.description}
              </p>
              <p className="text-sm">
                {tool.type} · {tool.enabled ? "Enabled" : "Disabled"} ·{" "}
                {tool.timeoutMs} ms
              </p>
            </li>
          ))}
        </ul>
      ) : (
        <p className="rounded-lg border border-dashed p-6 text-sm text-muted-foreground">
          No agent tools to display. Tool data is not connected.
        </p>
      )}
      <p className="text-sm text-muted-foreground">
        UI preview. Tool execution is not connected.
      </p>
    </section>
  );
}
