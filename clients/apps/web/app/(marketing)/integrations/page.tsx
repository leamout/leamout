import { AIProviders } from "@/components/integrations/ai-providers";
import { IntegrationsIntro } from "@/components/integrations/intro";
import { IntegrationsTelephony } from "@/components/integrations/telephony";
import { ToolsAndEvents } from "@/components/integrations/tools-and-events";
import { JsonLd } from "@/components/seo/json-ld";
import { constructMetadata } from "@/lib/metadata";
import { breadcrumbsSchema } from "@/lib/seo/breadcrumbs";
import { organizationSchema } from "@/lib/seo/organization";
import { webpageSchema } from "@/lib/seo/webpage";
import { websiteSchema } from "@/lib/seo/website";

export const metadata = constructMetadata({
  title: "Integrations",
  description:
    "Configure SIP trunks, verify AI integrations, assign provider roles, and connect tools and webhooks to Leamout Voice Agents.",
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
