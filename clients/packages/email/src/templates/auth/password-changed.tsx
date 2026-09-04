import { Button, Section, Text } from "react-email";
import { EmailLayout } from "../../components/layout";

type PasswordChangedEmailProps = {
  changedAt?: string;
  securityUrl?: string;
};

export default function PasswordChangedEmail({
  changedAt = "{{.ChangedAt}}",
  securityUrl = "{{.SecurityURL}}",
}: PasswordChangedEmailProps) {
  return (
    <EmailLayout preview="Your Leamout password was changed">
      <Section className="px-6 pt-8 pb-10 sm:px-10 sm:pt-10 sm:pb-12">
        <Text className="m-0 font-sans text-[40px] leading-[1.05] tracking-[-1.2px] text-fg">
          Password changed
        </Text>
        <Text className="m-0 mt-[18px] max-w-[470px] font-sans text-[14px] leading-[21px] text-fg-2">
          Your Leamout account password was changed at {changedAt}.
        </Text>
        <Text className="m-0 mt-[18px] max-w-[470px] font-sans text-[14px] leading-[21px] text-fg-2">
          If you made this change, no action is required. If you didn&apos;t,
          review your account security immediately.
        </Text>
        <Button
          href={securityUrl}
          className="mt-9 inline-block bg-brand px-5 py-3.5 text-center font-sans text-[15px] font-medium text-fg-inverted"
        >
          Review account security
        </Button>
      </Section>
    </EmailLayout>
  );
}

PasswordChangedEmail.PreviewProps = {
  changedAt: "September 4, 2026 at 1:30 PM UTC",
  securityUrl: "https://console.leamout.com/settings/security",
} satisfies PasswordChangedEmailProps;
