import { Heading, Section, Text } from "react-email";
import { EmailLayout } from "../../components/layout";

type AuthOTPEmailProps = {
  code?: string;
  expiresMinutes?: string;
};

export default function AuthOTPEmail({
  code = "{{.Code}}",
  expiresMinutes = "{{.ExpiresMinutes}}",
}: AuthOTPEmailProps) {
  return (
    <EmailLayout preview="Your Leamout verification code">
      <Heading as="h1" className="m-0 mb-4 text-2xl font-bold leading-8 text-neutral-900">
        Sign in to Leamout
      </Heading>
      <Text className="m-0 mb-5 text-[15px] leading-6 text-neutral-700">
        Use this verification code to continue signing in.
      </Text>
      <Section className="my-6 rounded-[10px] border border-neutral-200 bg-neutral-100 p-5 text-center">
        <Text className="m-0 font-mono text-[32px] font-bold leading-10 tracking-[0.18em] text-neutral-900">
          {code}
        </Text>
      </Section>
      <Text className="m-0 mb-5 text-[15px] leading-6 text-neutral-700">
        This code expires in {expiresMinutes} minutes.
      </Text>
      <Text className="m-0 text-[13px] leading-5 text-neutral-500">
        If you did not request this code, you can ignore this email.
      </Text>
    </EmailLayout>
  );
}
