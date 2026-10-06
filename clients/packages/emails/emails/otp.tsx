import { Heading, Text } from "react-email";
import Layout from "./_components/layout";

type OTPProps = {
  code: string;
  expiresAt: string;
};

export default function OTPEmail({ code, expiresAt }: OTPProps) {
  return (
    <Layout preview="Your Leamout sign-in code">
      <Heading as="h1" style={{ fontSize: "22px" }}>
        Your sign-in code
      </Heading>
      <Text>Enter this code to continue signing in to Leamout:</Text>
      <Text
        style={{ fontSize: "32px", fontWeight: "bold", letterSpacing: "6px" }}
      >
        {code}
      </Text>
      <Text>Expires at {expiresAt}.</Text>
      <Text>
        Do not share this code with anyone. If you did not request it, you can
        ignore this email.
      </Text>
    </Layout>
  );
}

OTPEmail.PreviewProps = {
  code: "012345",
  expiresAt: "12:30 UTC on 06 Oct 2026",
} satisfies OTPProps;
