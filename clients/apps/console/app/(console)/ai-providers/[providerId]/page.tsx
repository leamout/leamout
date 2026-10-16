import { IntegrationEditor } from "@/components/resources/integration-editor";
export default async function Page({params}:PageProps<"/ai-providers/[providerId]">){const {providerId}=await params;return <IntegrationEditor provider={providerId}/>;}
