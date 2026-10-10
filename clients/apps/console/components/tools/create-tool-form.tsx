import Link from "next/link";
import { ToolForm } from "@/components/tools/tool-form";

export function CreateToolForm() {
  return (
    <div className="space-y-6">
      <Link
        href="/tools"
        className="text-sm text-muted-foreground hover:underline"
      >
        Back to tools
      </Link>
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">Create tool</h1>
        <p className="text-muted-foreground">
          Define an action and the agent that owns it.
        </p>
      </div>
      <ToolForm creating />
    </div>
  );
}
