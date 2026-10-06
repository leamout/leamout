import type { ReactNode } from "react";
import { Body, Container, Head, Hr, Html, Preview, Text } from "react-email";

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
        style={{ backgroundColor: "#f4f5f7", fontFamily: "Arial, sans-serif" }}
      >
        <Container
          style={{
            backgroundColor: "#ffffff",
            borderRadius: "8px",
            margin: "40px auto",
            maxWidth: "560px",
            padding: "32px",
          }}
        >
          <Text
            style={{ color: "#111827", fontSize: "24px", fontWeight: "bold" }}
          >
            Leamout
          </Text>
          {children}
          <Hr style={{ borderColor: "#e5e7eb", marginTop: "32px" }} />
          <Text
            style={{ color: "#6b7280", fontSize: "12px", lineHeight: "20px" }}
          >
            Sent by Leamout. Never share your sign-in code or invitation link.
          </Text>
        </Container>
      </Body>
    </Html>
  );
}
