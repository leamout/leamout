import Link from "next/link";
import { ToolForm } from "@/components/tools/tool-form";
import type { Tool } from "@/components/tools/types";

export function ToolDetails({ toolId, tool }: { toolId: string; tool?: Tool }) {
  return (
    <div className="space-y-8">
      <Link
        href="/tools"
        className="text-sm text-muted-foreground hover:underline"
      >
        Back to tools
      </Link>
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">
          {tool?.name ?? "Tool details"}
        </h1>
        <p className="break-all text-sm text-muted-foreground">
          Tool ID: {toolId}
        </p>
        <p className="text-sm text-muted-foreground">
          UI preview.{" "}
          {tool
            ? "Preview configuration is shown below."
            : "Tool configuration has not been loaded."}
        </p>
      </div>
      <section className="space-y-2 rounded-lg border p-6">
        <h2 className="font-semibold">Owning agent</h2>
        {tool ? (
          <Link
            className="text-sm underline"
            href={`/agents/${encodeURIComponent(tool.agent.id)}`}
          >
            {tool.agent.name}
          </Link>
        ) : (
          <p className="text-sm text-muted-foreground">
            Agent ownership has not been loaded.
          </p>
        )}
      </section>
      <ToolForm tool={tool} />
      <section className="space-y-2 rounded-lg border p-6">
        <h2 className="font-semibold">Execution and signing</h2>
        <p className="text-sm text-muted-foreground">
          Tool executions are reviewed within call details. Signing-secret
          rotation will be connected later; no secret is displayed or generated
          in this preview.
        </p>
        <Link href="/calls" className="text-sm underline">
          View calls
        </Link>
      </section>
    </div>
  );
}
