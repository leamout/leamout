import { Button } from "@leamout/ui/components/button";
import Link from "next/link";
import Pricing from "@/components/shadcn-space/blocks/pricing-01/pricing";
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
      <Pricing signupHref={`${consoleURL}/create-account`} />
      <div className="mx-auto max-w-6xl px-5 sm:px-6 lg:px-8">
        <section className="my-6 flex flex-col justify-between gap-6 rounded-xl bg-muted/40 p-8 sm:flex-row sm:items-center">
          <div>
            <h2 className="text-2xl font-semibold">Enterprise</h2>
            <p className="mt-2 max-w-2xl leading-7 text-muted-foreground">
              Custom capacity, SSO, SCIM, private networking, and deployment
              requirements. Scope and pricing are agreed for your organization.
            </p>
          </div>
          <Button
            variant="outline"
            nativeButton={false}
            render={<Link href="/contact" />}
          >
            Contact us
          </Button>
        </section>
        <section className="py-16">
          <h2 className="text-3xl font-semibold tracking-tight">
            Know what you are paying for.
          </h2>
          <div className="mt-8 grid gap-8 md:grid-cols-3">
            {[
              [
                "Leamout",
                "Your platform plan includes runtime capacity, agent and trunk limits, and retention. Pro and Scale also include advanced roles and configurable retention policies.",
              ],
              [
                "Your carrier",
                "You bring existing numbers and SIP connectivity. Your carrier bills you for phone numbers and telephony usage.",
              ],
              [
                "Your AI providers",
                "You supply your own provider credentials. Your AI providers bill you for speech and model usage.",
              ],
            ].map(([title, text]) => (
              <div key={title} className="border-t border-border pt-5">
                <h3 className="font-semibold">{title}</h3>
                <p className="mt-3 leading-7 text-muted-foreground">{text}</p>
              </div>
            ))}
          </div>
        </section>
        <section className="border-t border-border py-12">
          <h2 className="text-2xl font-semibold">Common questions</h2>
          {[
            [
              "Are carrier and AI costs included?",
              "No. Leamout does not resell carrier minutes or AI usage. Those charges remain with the providers you choose.",
            ],
            [
              "What happens when my runtime allowance is exhausted?",
              "New voice-agent sessions are rejected once the monthly allowance is exhausted. Active calls can finish; there is no automatic usage top-up.",
            ],
            [
              "Can I use my existing numbers?",
              "Yes. Connect them through your SIP carrier, PBX, or SBC. Number purchasing is not part of Leamout.",
            ],
            [
              "Can I self-host?",
              "Leamout supports self-hosted deployment. Read the documentation for setup requirements and contact us to discuss enterprise deployment support.",
            ],
          ].map(([question, answer]) => (
            <details key={question} className="border-b border-border py-5">
              <summary className="cursor-pointer font-medium focus-visible:outline-brand-orange">
                {question}
              </summary>
              <p className="mt-4 max-w-3xl leading-7 text-muted-foreground">
                {answer}
              </p>
            </details>
          ))}
        </section>
      </div>
    </>
  );
}
