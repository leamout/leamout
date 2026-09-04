import { Button, Section, Text } from "react-email";
import { EmailLayout } from "../../components/layout";

type BackupFailedEmailProps = {
  deploymentName?: string;
  failedAt?: string;
  errorSummary?: string;
  deploymentUrl?: string;
};

export default function BackupFailedEmail({
  deploymentName = "{{.DeploymentName}}",
  failedAt = "{{.FailedAt}}",
  errorSummary = "{{.ErrorSummary}}",
  deploymentUrl = "{{.DeploymentURL}}",
}: BackupFailedEmailProps) {
  return (
    <EmailLayout preview={`Backup failed for ${deploymentName}`}>
      <Section className="px-6 pt-8 pb-10 sm:px-10 sm:pt-10 sm:pb-12">
        <Text className="m-0 font-sans text-[40px] leading-[1.05] tracking-[-1.2px] text-fg">
          Backup failed
        </Text>
        <Text className="m-0 mt-[18px] max-w-[470px] font-sans text-[14px] leading-[21px] text-fg-2">
          Leamout couldn&apos;t complete the scheduled backup for{" "}
          {deploymentName} at {failedAt}.
        </Text>
        <Section className="mt-8 rounded-[6px] border border-[#F0C9C2] bg-danger-muted px-5 py-4">
          <Text className="m-0 font-sans text-[12px] font-medium uppercase tracking-[0.06em] text-danger">
            Failure details
          </Text>
          <Text className="m-0 mt-2 font-sans text-[13px] leading-[20px] text-danger">
            {errorSummary}
          </Text>
        </Section>
        <Button
          href={deploymentUrl}
          className="mt-9 inline-block bg-brand px-5 py-3.5 text-center font-sans text-[15px] font-medium text-fg-inverted"
        >
          Review deployment
        </Button>
      </Section>
    </EmailLayout>
  );
}

BackupFailedEmail.PreviewProps = {
  deploymentName: "production-voice",
  failedAt: "September 4, 2026 at 02:00 UTC",
  errorSummary: "Backup archive upload did not complete before the timeout.",
  deploymentUrl: "https://console.leamout.com/deployments/production-voice",
} satisfies BackupFailedEmailProps;
