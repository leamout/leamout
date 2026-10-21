import { PlatformArchitecture } from "@/components/platform/architecture";
import { PlatformCapabilities } from "@/components/platform/capabilities";
import { PlatformIntro } from "@/components/platform/intro";
import { JsonLd } from "@/components/seo/json-ld";
import { constructMetadata } from "@/lib/metadata";
import { breadcrumbsSchema } from "@/lib/seo/breadcrumbs";
import { organizationSchema } from "@/lib/seo/organization";
import { webpageSchema } from "@/lib/seo/webpage";
import { websiteSchema } from "@/lib/seo/website";

export const metadata = constructMetadata({
  title: "Platform",
  description:
    "Explore the managed Leamout platform for SIP-native calling, realtime voice agents, call control, and AI orchestration.",
  path: "/platform",
});

export default function PlatformPage() {
  return (
    <>
      <JsonLd
        nodes={[
          organizationSchema(),
          websiteSchema(),
          webpageSchema({
            path: "/platform",
            name: "Leamout Platform",
            description:
              "Explore Leamout's SIP-native runtime, realtime voice agents, call control, and AI orchestration.",
          }),
          breadcrumbsSchema([
            { name: "Home", path: "/" },
            { name: "Platform", path: "/platform" },
          ]),
        ]}
      />
      <PlatformIntro />
      <PlatformArchitecture />
      <PlatformCapabilities />
    </>
  );
}
