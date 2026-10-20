import { PlatformArchitecture } from "@/components/platform/architecture";
import { PlatformCapabilities } from "@/components/platform/capabilities";
import { PlatformIntro } from "@/components/platform/intro";
import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Platform",
  description:
    "Explore the managed Leamout platform for SIP-native calling, realtime voice agents, call control, and AI orchestration.",
  path: "/platform",
});

export default function PlatformPage() {
  return (
    <>
      <PlatformIntro />
      <PlatformArchitecture />
      <PlatformCapabilities />
    </>
  );
}
