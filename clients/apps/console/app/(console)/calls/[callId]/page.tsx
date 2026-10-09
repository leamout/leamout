import { CallDetails } from "@/components/calls/call-details";

export default async function Page({ params }: PageProps<"/calls/[callId]">) {
  const { callId } = await params;

  return <CallDetails callId={callId} />;
}
