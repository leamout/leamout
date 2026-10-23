"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { ToolForm } from "./tool-form";

export function CreateToolForm() {
  const agentId = useSearchParams().get("agent_id") ?? "";
  return (
    <div className="space-y-6">
      <Link href="/tools" className="text-sm text-muted-foreground hover:underline">Back to tools</Link>
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold">Create tool</h1>
        <p className="text-muted-foreground">Define a tool owned by one Voice Agent.</p>
      </div>
      <ToolForm creating initialAgentId={agentId} />
    </div>
  );
}
