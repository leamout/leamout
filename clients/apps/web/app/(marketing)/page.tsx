import { Capabilities } from "@/components/home/capabilities";
import { constructMetadata } from "@/lib/metadata";
import { ControlPlane } from "@/components/home/control-plane";
import { HomeCallToAction } from "@/components/home/call-to-action";
import { HomeHero } from "@/components/home/hero";
import { HowItWorks } from "@/components/home/how-it-works";
import { JsonLd } from "@/components/seo/json-ld";
import { organizationSchema } from "@/lib/seo/organization";
import { Ownership } from "@/components/home/ownership";
import { softwareSchema } from "@/lib/seo/software";
import { webpageSchema } from "@/lib/seo/webpage";
import { websiteSchema } from "@/lib/seo/website";

export const metadata = constructMetadata({
  title: "SIP-Native Voice Agent Runtime & Control Plane",
  description:
    "Leamout is a SIP-native runtime and control plane for autonomous voice agents, with programmable call control, real-time media, and AI orchestration.",
  path: "/",
});

export default function Home() {
  return (
    <>
      <JsonLd
        nodes={[
          organizationSchema(),
          websiteSchema(),
          webpageSchema({
            path: "/",
            name: "Leamout — SIP-Native Voice Agent Runtime & Control Plane",
            description:
              "SIP-native runtime and control plane for autonomous voice agents, programmable call control, real-time media, and AI orchestration.",
          }),
          softwareSchema(),
        ]}
      />
      <HomeHero />
      <HowItWorks />
      <Capabilities />
      <ControlPlane />
      <Ownership />
      <HomeCallToAction />
    </>
  );
}
