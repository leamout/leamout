import { Button, Section, Text } from "react-email";
import { EmailLayout } from "../../components/layout";

type PasswordResetEmailProps = {
  resetUrl?: string;
  expiresMinutes?: string;
};

export default function PasswordResetEmail({
  resetUrl = "{{.ResetURL}}",
  expiresMinutes = "{{.ExpiresMinutes}}"
}: PasswordResetEmailProps) {
  return (
    <EmailLayout preview="Reset your Leamout password">
      <Section className="px-6 pt-8 pb-10 sm:px-10 sm:pt-10 sm:pb-12">
        <Text className="m-0 font-sans text-[48px] leading-none tracking-[-1.44px] text-fg">
          Reset your password
        </Text>
        <Text className="m-0 mt-[18px] max-w-[470px] font-sans text-[14px] leading-[21px] text-fg-2">
          Someone requested a password reset for your Leamout account. Use the button below to choose a new password.
        </Text>
        <Button
          href={resetUrl}
          className="mt-9 inline-block bg-brand px-5 py-3.5 text-center font-sans text-[15px] font-medium text-fg-inverted"
        >
          Change password
        </Button>
        <Text className="m-0 mt-12 max-w-[390px] font-sans text-[11px] leading-[17px] text-fg-3">
          This link expires in {expiresMinutes} minutes. If you didn&apos;t request this, ignore this email. Your password won&apos;t change unless the reset link is used.
        </Text>
      </Section>
    </EmailLayout>
  );
}

PasswordResetEmail.PreviewProps = {
  resetUrl: "https://console.leamout.com/auth/reset-password?token=preview",
  expiresMinutes: "30"
} satisfies PasswordResetEmailProps;
