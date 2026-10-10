import { RecordingDetails } from "@/components/recordings/recording-details";

export default async function Page({
  params,
}: PageProps<"/recordings/[recordingId]">) {
  const { recordingId } = await params;

  return <RecordingDetails key={recordingId} recordingId={recordingId} />;
}
