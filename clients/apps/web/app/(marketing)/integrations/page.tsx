import { AIProviders } from "@/components/integrations/ai-providers";
import { breadcrumbsSchema } from "@/lib/seo/breadcrumbs";
import { constructMetadata } from "@/lib/metadata";
import { IntegrationsIntro } from "@/components/integrations/intro";
import { IntegrationsTelephony } from "@/components/integrations/telephony";
import { JsonLd } from "@/components/seo/json-ld";
import { organizationSchema } from "@/lib/seo/organization";
import { ToolsAndEvents } from "@/components/integrations/tools-and-events";
import { webpageSchema } from "@/lib/seo/webpage";
import { websiteSchema } from "@/lib/seo/website";

export const metadata = constructMetadata({
  title: "Integrations",
  description:
    "Connect SIP trunks, AI providers, tools, and webhooks to autonomous voice agents on Leamout.",
  path: "/integrations",
});

export default function IntegrationsPage() {
  return (
    <>
      <JsonLd
        nodes={[
          organizationSchema(),
          websiteSchema(),
          webpageSchema({
            path: "/integrations",
            name: "Leamout Integrations",
            description:
              "Connect SIP trunks, AI providers, tools, and webhooks to autonomous voice agents on Leamout.",
          }),
          breadcrumbsSchema([
            { name: "Home", path: "/" },
            { name: "Integrations", path: "/integrations" },
          ]),
        ]}
      />
      <IntegrationsIntro />
      <IntegrationsTelephony />
      <AIProviders />
      <ToolsAndEvents />
    </>
  );
}
