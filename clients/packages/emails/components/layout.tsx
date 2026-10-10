import type { CSSProperties, ReactNode } from "react";
import {
  Body,
  Container,
  Head,
  Html,
  Preview,
  Section,
  Text,
} from "react-email";

export const headingStyle: CSSProperties = {
  color: "#171717",
  fontSize: "26px",
  fontWeight: "600",
  lineHeight: "1.3",
  margin: "0 0 20px",
};

export const textStyle: CSSProperties = {
  color: "#404040",
  fontSize: "15px",
  lineHeight: "1.7",
  margin: "16px 0",
};

export const buttonStyle: CSSProperties = {
  backgroundColor: "#171717",
  borderRadius: "6px",
  color: "#ffffff",
  fontSize: "14px",
  fontWeight: "600",
  padding: "14px 22px",
  textDecoration: "none",
};

export default function Layout({
  preview,
  children,
}: {
  preview: string;
  children: ReactNode;
}) {
  return (
    <Html lang="en">
      <Head />
      <Preview>{preview}</Preview>
      <Body
        style={{
          backgroundColor: "#f5f5f5",
          fontFamily: "Arial, Helvetica, sans-serif",
          margin: "0",
          padding: "32px 12px",
        }}
      >
        <Container style={{ maxWidth: "560px", margin: "0 auto" }}>
          <Section style={{ padding: "0 8px 20px" }}>
            <Text
              style={{
                color: "#171717",
                fontSize: "20px",
                fontWeight: "700",
                letterSpacing: "-0.5px",
                margin: "0",
              }}
            >
              Leamout
            </Text>
          </Section>
          <Section
            style={{
              backgroundColor: "#ffffff",
              border: "1px solid #e5e5e5",
              borderRadius: "8px",
              padding: "32px 24px",
            }}
          >
            {children}
          </Section>
          <Section style={{ padding: "20px 8px" }}>
            <Text
              style={{
                color: "#737373",
                fontSize: "12px",
                lineHeight: "1.6",
                margin: "0",
              }}
            >
              Leamout · Autonomous voice agents
            </Text>
          </Section>
        </Container>
      </Body>
    </Html>
  );
}
