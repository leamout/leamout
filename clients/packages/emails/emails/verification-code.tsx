import { Heading, Text } from "react-email";
import Layout, { headingStyle, textStyle } from "../components/layout";

type VerificationCodeProps = {
  code: string;
  expiresInMinutes?: number;
  expiresAt?: string;
  purpose: "create-account" | "log-in" | "password-recovery";
};
const messages = {
  "create-account": {
    title: "Verify your email",
    body: "Use this code to verify your email address and continue creating your Leamout account.",
  },
  "log-in": {
    title: "Your login code",
    body: "Use this code to continue logging in to Leamout.",
  },
  "password-recovery": {
    title: "Verify your recovery request",
    body: "Use this code to continue resetting your Leamout password.",
  },
};

export default function VerificationCode({
  code,
  expiresInMinutes,
  expiresAt,
  purpose,
}: VerificationCodeProps) {
  const message = messages[purpose];
  return (
    <Layout preview={message.title}>
      <Heading style={headingStyle}>{message.title}</Heading>
      <Text style={textStyle}>{message.body}</Text>
      <Text
        style={{
          backgroundColor: "#f5f5f5",
          borderRadius: "6px",
          color: "#171717",
          fontFamily: "monospace",
          fontSize: "32px",
          fontWeight: "700",
          letterSpacing: "6px",
          padding: "20px",
          textAlign: "center",
        }}
      >
        {code}
      </Text>
      <Text style={textStyle}>
        {expiresAt
          ? `This code expires ${expiresAt}.`
          : `This code expires in ${expiresInMinutes} minutes.`}{" "}
        Do not share it with anyone.
      </Text>
      <Text style={textStyle}>
        If you did not request this code, you can ignore this email.
      </Text>
    </Layout>
  );
}

VerificationCode.PreviewProps = {
  code: "482916",
  expiresInMinutes: 10,
  purpose: "create-account",
} satisfies VerificationCodeProps;

VerificationCode.ExportProps = {
  code: "{{.Code}}",
  expiresAt: "{{.Expiry}}",
  purpose: "log-in",
} satisfies VerificationCodeProps;
