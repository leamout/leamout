import { TrunkEditor } from "@/components/resources/trunk-editor";
export default async function Page({params}:PageProps<"/sip-trunks/[trunkId]">){const {trunkId}=await params;return <TrunkEditor id={trunkId}/>;}
