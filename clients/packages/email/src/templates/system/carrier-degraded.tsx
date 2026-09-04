import { Button, Section, Text } from "react-email";
import { EmailLayout } from "../../components/layout";

type CarrierDegradedEmailProps = {
  carrierName?: string;
  connectionName?: string;
  detectedAt?: string;
  connectionUrl?: string;
};

export default function CarrierDegradedEmail({
  carrierName = "{{.CarrierName}}",
  connectionName = "{{.ConnectionName}}",
  detectedAt = "{{.DetectedAt}}",
  connectionUrl = "{{.ConnectionURL}}",
}: CarrierDegradedEmailProps) {
  return (
    <EmailLayout preview={`${carrierName} connection degraded`}>
      <Section className="px-6 pt-8 pb-10 sm:px-10 sm:pt-10 sm:pb-12">
        <Text className="m-0 font-sans text-[40px] leading-[1.05] tracking-[-1.2px] text-fg">
          Carrier connection degraded
        </Text>
        <Text className="m-0 mt-[18px] max-w-[470px] font-sans text-[14px] leading-[21px] text-fg-2">
          Leamout detected degraded health for {connectionName} on {carrierName}{" "}
          at {detectedAt}.
        </Text>
        <Text className="m-0 mt-[18px] max-w-[470px] font-sans text-[14px] leading-[21px] text-fg-2">
          Calls may fail over to another configured endpoint while the
          connection recovers.
        </Text>
        <Button
          href={connectionUrl}
          className="mt-9 inline-block bg-brand px-5 py-3.5 text-center font-sans text-[15px] font-medium text-fg-inverted"
        >
          Review carrier connection
        </Button>
      </Section>
    </EmailLayout>
  );
}

CarrierDegradedEmail.PreviewProps = {
  carrierName: "DIDWW",
  connectionName: "didww-primary",
  detectedAt: "September 4, 2026 at 13:45 UTC",
  connectionUrl: "https://console.leamout.com/carriers/didww-primary",
} satisfies CarrierDegradedEmailProps;
