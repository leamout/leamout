import { ToolDetails } from "@/components/tools/tool-details";

export default async function Page({ params }: PageProps<"/tools/[toolId]">) {
  const { toolId } = await params;

  return <ToolDetails key={toolId} toolId={toolId} />;
}
