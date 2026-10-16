import { ToolEditor } from "@/components/resources/tool-editor";
export default async function Page({params}:PageProps<"/tools/[toolId]">){const {toolId}=await params;return <ToolEditor id={toolId}/>;}
