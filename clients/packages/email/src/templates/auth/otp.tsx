import { Section, Text } from "react-email";
import { EmailLayout } from "../../components/layout";

type AuthOTPEmailProps = {
  code?: string;
  expiresMinutes?: string;
};

export default function AuthOTPEmail({
  code = "{{.Code}}",
  expiresMinutes = "{{.ExpiresMinutes}}"
}: AuthOTPEmailProps) {
  return (
    <EmailLayout preview="Your Leamout verification code">
      <Section className="px-6 pt-8 pb-10 sm:px-10 sm:pt-10 sm:pb-12">
        <Text className="m-0 font-sans text-[40px] leading-[1.05] tracking-[-1.2px] text-fg">
          Your verification code
        </Text>
        <Text className="m-0 mt-[18px] font-sans text-[14px] leading-[21px] text-fg-2">
          Use this code to continue signing in to Leamout.
        </Text>
        <Section className="my-8 rounded-[6px] border border-stroke bg-bg-muted px-5 py-5 text-center">
          <Text className="m-0 font-mono text-[32px] font-medium leading-[40px] tracking-[0.18em] text-fg">
            {code}
          </Text>
        </Section>
        <Text className="m-0 font-sans text-[13px] leading-[20px] text-fg-3">
          This code expires in {expiresMinutes} minutes. If you didn&apos;t request it, you can safely ignore this email.
        </Text>
      </Section>
    </EmailLayout>
  );
}

AuthOTPEmail.PreviewProps = {
  code: "482913",
  expiresMinutes: "10"
} satisfies AuthOTPEmailProps;
