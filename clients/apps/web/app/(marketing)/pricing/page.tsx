import { BillingExplanation } from "@/components/pricing/billing-explanation";
import { Enterprise } from "@/components/pricing/enterprise";
import { PricingFAQ } from "@/components/pricing/faq";
import PricingPlans from "@/components/pricing/pricing-plans";
import { constructMetadata } from "@/lib/metadata";
import { consoleURL } from "@/lib/site";

export const metadata = constructMetadata({
  title: "Pricing",
  description:
    "Explore Leamout platform subscriptions for running autonomous voice agents with your own carriers and AI providers.",
  path: "/pricing",
});

export default function PricingPage() {
  return (
    <>
      <PricingPlans signupHref={`${consoleURL}/create-account`} />
      <div className="mx-auto max-w-6xl px-5 sm:px-6 lg:px-8">
        <Enterprise />
        <BillingExplanation />
        <PricingFAQ />
      </div>
    </>
  );
}
