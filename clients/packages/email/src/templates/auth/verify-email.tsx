import { Button, Section, Text } from "react-email";
import { EmailLayout } from "../../components/layout";

type VerifyEmailProps = {
  verifyUrl?: string;
};

export default function VerifyEmail({ verifyUrl = "{{.VerifyURL}}" }: VerifyEmailProps) {
  return (
    <EmailLayout preview="Confirm your Leamout email address">
      <Section className="px-6 pt-8 pb-10 sm:px-10 sm:pt-10 sm:pb-12">
        <Text className="m-0 font-sans text-[48px] leading-none tracking-[-1.44px] text-fg">
          Almost there
        </Text>
        <Text className="m-0 mt-[18px] max-w-[460px] font-sans text-[14px] leading-[21px] text-fg-2">
          Thanks for creating your Leamout account. Confirm your email address to finish setting it up.
        </Text>
        <Button
          href={verifyUrl}
          className="mt-9 inline-block bg-brand px-5 py-3.5 text-center font-sans text-[15px] font-medium text-fg-inverted"
        >
          Confirm email
        </Button>
        <Text className="m-0 mt-12 max-w-[340px] font-sans text-[11px] leading-[17px] text-fg-3">
          If you didn&apos;t create a Leamout account, you can safely ignore this email.
        </Text>
      </Section>
    </EmailLayout>
  );
}

VerifyEmail.PreviewProps = {
  verifyUrl: "https://console.leamout.com/auth/verify?token=preview"
} satisfies VerifyEmailProps;
