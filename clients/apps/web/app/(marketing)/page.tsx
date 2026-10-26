import { HomeCallToAction } from "@/components/home/call-to-action";
import { ControlPlane } from "@/components/home/control-plane";
import { DeveloperQuickstart } from "@/components/home/developer-quickstart";
import { HomeHero } from "@/components/home/hero";
import { HowItWorks } from "@/components/home/how-it-works";
import { JsonLd } from "@/components/seo/json-ld";
import { organizationSchema } from "@/lib/seo/organization";
import { founderSchema } from "@/lib/seo/person";
import { softwareSchema } from "@/lib/seo/software";
import { webpageSchema } from "@/lib/seo/webpage";
import { websiteSchema } from "@/lib/seo/website";

export default function Home() {
  return (
    <>
      <JsonLd
        nodes={[
          organizationSchema(),
          founderSchema(),
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
      <ControlPlane />
      <DeveloperQuickstart />
      <HomeCallToAction />
    </>
  );
}
