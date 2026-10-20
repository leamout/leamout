import { BillingExplanation } from "@/components/pricing/billing-explanation";
import { PricingFAQ } from "@/components/pricing/faq";
import PricingPlans from "@/components/pricing/pricing-plans";
import { constructMetadata } from "@/lib/metadata";
import { consoleURL } from "@/lib/site";

export const metadata = constructMetadata({
  title: "Pricing",
  description:
    "Compare Leamout subscriptions: runtime minutes, concurrent calls, voice agents, SIP trunks, and retention.",
  path: "/pricing",
});

export default function PricingPage() {
  return (
    <>
      <PricingPlans signupHref={consoleURL + "/create-account"} />
      <div className="mx-auto max-w-[1200px] px-5 sm:px-6 lg:px-8">
        <BillingExplanation />
        <PricingFAQ />
      </div>
    </>
  );
}
