import { Button, Section, Text } from "react-email";
import { EmailLayout } from "../../components/layout";

type DeploymentAlertEmailProps = {
  deploymentName?: string;
  alertTitle?: string;
  alertMessage?: string;
  occurredAt?: string;
  deploymentUrl?: string;
};

export default function DeploymentAlertEmail({
  deploymentName = "{{.DeploymentName}}",
  alertTitle = "{{.AlertTitle}}",
  alertMessage = "{{.AlertMessage}}",
  occurredAt = "{{.OccurredAt}}",
  deploymentUrl = "{{.DeploymentURL}}",
}: DeploymentAlertEmailProps) {
  return (
    <EmailLayout preview={`${alertTitle} on ${deploymentName}`}>
      <Section className="px-6 pt-8 pb-10 sm:px-10 sm:pt-10 sm:pb-12">
        <Text className="m-0 font-sans text-[40px] leading-[1.05] tracking-[-1.2px] text-fg">
          {alertTitle}
        </Text>
        <Text className="m-0 mt-[18px] max-w-[470px] font-sans text-[14px] leading-[21px] text-fg-2">
          Leamout detected an issue on {deploymentName} at {occurredAt}.
        </Text>
        <Section className="mt-8 rounded-[6px] border border-[#EAD79B] bg-warning-muted px-5 py-4">
          <Text className="m-0 font-sans text-[13px] leading-[20px] text-warning">
            {alertMessage}
          </Text>
        </Section>
        <Button
          href={deploymentUrl}
          className="mt-9 inline-block bg-brand px-5 py-3.5 text-center font-sans text-[15px] font-medium text-fg-inverted"
        >
          Inspect deployment
        </Button>
      </Section>
    </EmailLayout>
  );
}

DeploymentAlertEmail.PreviewProps = {
  deploymentName: "production-voice",
  alertTitle: "Worker health degraded",
  alertMessage:
    "One or more worker supervisors have been unhealthy for five minutes.",
  occurredAt: "September 4, 2026 at 13:42 UTC",
  deploymentUrl: "https://console.leamout.com/deployments/production-voice",
} satisfies DeploymentAlertEmailProps;
