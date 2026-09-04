import { Button, Link, Section, Text } from "react-email";
import { EmailLayout } from "../../components/layout";

type FeatureAnnouncementEmailProps = {
  featureName?: string;
  summary?: string;
  url?: string;
  unsubscribeUrl?: string;
};

export default function FeatureAnnouncementEmail({
  featureName = "{{.FeatureName}}",
  summary = "{{.Summary}}",
  url = "{{.URL}}",
  unsubscribeUrl = "{{.UnsubscribeURL}}",
}: FeatureAnnouncementEmailProps) {
  return (
    <EmailLayout preview={`${featureName} is now available in Leamout`}>
      <Section className="px-6 pt-8 pb-10 sm:px-10 sm:pt-10 sm:pb-12">
        <Text className="m-0 font-sans text-[48px] leading-none tracking-[-1.44px] text-fg">
          Meet {featureName}
        </Text>
        <Text className="m-0 mt-[18px] max-w-[480px] font-sans text-[14px] leading-[21px] text-fg-2">
          {summary}
        </Text>
        <Button
          href={url}
          className="mt-9 inline-block bg-brand px-5 py-3.5 text-center font-sans text-[15px] font-medium text-fg-inverted"
        >
          See what&apos;s new
        </Button>
        <Text className="m-0 mt-12 font-sans text-[11px] leading-[17px] text-fg-3">
          <Link href={unsubscribeUrl} className="text-fg-3 underline">
            Unsubscribe
          </Link>{" "}
          from Leamout product emails.
        </Text>
      </Section>
    </EmailLayout>
  );
}

FeatureAnnouncementEmail.PreviewProps = {
  featureName: "Carrier health",
  summary:
    "See carrier connection health and endpoint failover state in one place, with clearer signals when your voice path needs attention.",
  url: "https://leamout.com/updates/carrier-health",
  unsubscribeUrl: "https://leamout.com/email/preferences",
} satisfies FeatureAnnouncementEmailProps;
