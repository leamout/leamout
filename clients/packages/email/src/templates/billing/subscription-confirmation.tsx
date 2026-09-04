import { Button, Column, Row, Section, Text } from "react-email";
import { EmailLayout } from "../../components/layout";

type SubscriptionConfirmationEmailProps = {
  userName?: string;
  planName?: string;
  planPrice?: string;
  cycleLabel?: string;
  nextBillingDate?: string;
  subtotal?: string;
  tax?: string;
  total?: string;
  manageUrl?: string;
};

export default function SubscriptionConfirmationEmail({
  userName = "{{.UserName}}",
  planName = "{{.PlanName}}",
  planPrice = "{{.PlanPrice}}",
  cycleLabel = "{{.CycleLabel}}",
  nextBillingDate = "{{.NextBillingDate}}",
  subtotal = "{{.Subtotal}}",
  tax = "{{.Tax}}",
  total = "{{.Total}}",
  manageUrl = "{{.ManageURL}}"
}: SubscriptionConfirmationEmailProps) {
  return (
    <EmailLayout preview={`Your Leamout ${planName} subscription is active`}>
      <Section className="px-6 pt-8 pb-10 sm:px-10 sm:pt-10 sm:pb-12">
        <Text className="m-0 font-sans text-[48px] leading-none tracking-[-1.44px] text-fg">
          Subscription confirmed
        </Text>
        <Text className="m-0 mt-[18px] max-w-[480px] font-sans text-[14px] leading-[21px] text-fg-2">
          Hi {userName}, your Leamout {planName} subscription is active. You&apos;re billed {planPrice} per {cycleLabel}; your next charge is on {nextBillingDate}.
        </Text>
        <Section className="mt-9">
          <Section className="mb-[3px] bg-bg-muted p-3">
            <Row>
              <Column><Text className="m-0 font-sans text-[14px] text-fg">Subtotal</Text></Column>
              <Column align="right"><Text className="m-0 font-sans text-[14px] text-fg">{subtotal}</Text></Column>
            </Row>
          </Section>
          <Section className="mb-[3px] bg-bg-muted p-3">
            <Row>
              <Column><Text className="m-0 font-sans text-[14px] text-fg">Tax</Text></Column>
              <Column align="right"><Text className="m-0 font-sans text-[14px] text-fg">{tax}</Text></Column>
            </Row>
          </Section>
          <Section className="bg-[#DDE7D8] p-3">
            <Row>
              <Column><Text className="m-0 font-sans text-[14px] font-medium text-fg">Total</Text></Column>
              <Column align="right"><Text className="m-0 font-sans text-[14px] font-medium text-fg">{total}</Text></Column>
            </Row>
          </Section>
        </Section>
        <Button
          href={manageUrl}
          className="mt-9 inline-block bg-brand px-5 py-3.5 text-center font-sans text-[15px] font-medium text-fg-inverted"
        >
          Manage subscription
        </Button>
      </Section>
    </EmailLayout>
  );
}

SubscriptionConfirmationEmail.PreviewProps = {
  userName: "Alex",
  planName: "Pro",
  planPrice: "$29",
  cycleLabel: "month",
  nextBillingDate: "October 4, 2026",
  subtotal: "$29.00",
  tax: "$0.00",
  total: "$29.00",
  manageUrl: "https://console.leamout.com/settings/billing"
} satisfies SubscriptionConfirmationEmailProps;
