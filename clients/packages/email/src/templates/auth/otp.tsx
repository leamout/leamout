import type { CSSProperties } from "react";
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
      <Heading as="h1" style={heading}>
        Sign in to Leamout
      </Heading>
      <Text style={paragraph}>Use this verification code to continue signing in.</Text>
      <Section style={codeBox}>
        <Text style={codeText}>{code}</Text>
      </Section>
      <Text style={paragraph}>This code expires in {expiresMinutes} minutes.</Text>
      <Text style={muted}>If you did not request this code, you can ignore this email.</Text>
    </EmailLayout>
  );
}

const heading: CSSProperties = {
  margin: "0 0 16px",
  color: "#171717",
  fontSize: "24px",
  fontWeight: 700,
  lineHeight: "32px",
};

const paragraph: CSSProperties = {
  margin: "0 0 20px",
  color: "#404040",
  fontSize: "15px",
  lineHeight: "24px",
};

const codeBox: CSSProperties = {
  margin: "24px 0",
  padding: "20px",
  backgroundColor: "#f5f5f5",
  border: "1px solid #e5e5e5",
  borderRadius: "10px",
  textAlign: "center",
};

const codeText: CSSProperties = {
  margin: 0,
  color: "#111111",
  fontFamily: "ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace",
  fontSize: "32px",
  fontWeight: 700,
  letterSpacing: "0.18em",
  lineHeight: "40px",
};

const muted: CSSProperties = {
  margin: 0,
  color: "#737373",
  fontSize: "13px",
  lineHeight: "20px",
};
