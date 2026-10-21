import { BillingExplanation } from "@/components/pricing/billing-explanation";
import { breadcrumbsSchema } from "@/lib/seo/breadcrumbs";
import { consoleURL } from "@/lib/site";
import { constructMetadata } from "@/lib/metadata";
import { JsonLd } from "@/components/seo/json-ld";
import { organizationSchema } from "@/lib/seo/organization";
import { PricingFAQ } from "@/components/pricing/faq";
import { webpageSchema } from "@/lib/seo/webpage";
import { websiteSchema } from "@/lib/seo/website";
import PricingPlans from "@/components/pricing/pricing-plans";

export const metadata = constructMetadata({
  title: "Pricing",
  description:
    "Compare Leamout subscriptions: runtime minutes, concurrent calls, voice agents, SIP trunks, and retention.",
  path: "/pricing",
});

export default function PricingPage() {
  return (
    <>
      <JsonLd
        nodes={[
          organizationSchema(),
          websiteSchema(),
          webpageSchema({
            path: "/pricing",
            name: "Leamout Pricing",
            description:
              "Compare Leamout subscriptions: runtime minutes, concurrent calls, voice agents, SIP trunks, and retention.",
          }),
          breadcrumbsSchema([
            { name: "Home", path: "/" },
            { name: "Pricing", path: "/pricing" },
          ]),
        ]}
      />
      <PricingPlans signupHref={consoleURL + "/create-account"} />
      <div className="mx-auto max-w-[1200px] px-5 sm:px-6 lg:px-8">
        <BillingExplanation />
        <PricingFAQ />
      </div>
    </>
  );
}
