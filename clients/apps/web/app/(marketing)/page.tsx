import { HomeStructuredData } from "@/components/home/structured-data";
import { constructMetadata } from "@/lib/metadata";
import { HomeCallToAction } from "@/components/home/call-to-action";
import { Capabilities } from "@/components/home/capabilities";
import { ControlPlane } from "@/components/home/control-plane";
import { HomeHero } from "@/components/home/hero";
import { HowItWorks } from "@/components/home/how-it-works";
import { Ownership } from "@/components/home/ownership";

export const metadata = constructMetadata({
  title: "SIP-Native Voice Agent Runtime & Control Plane",
  description:
    "Leamout is a SIP-native runtime and control plane for autonomous voice agents, with programmable call control, real-time media, and AI orchestration.",
  path: "/",
});

export default function Home() {
  return (
    <>
      <HomeStructuredData />
      <HomeHero />
      <HowItWorks />
      <Capabilities />
      <ControlPlane />
      <Ownership />
      <HomeCallToAction />
    </>
  );
}
