import type { CSSProperties, ReactNode } from "react";
import { Body, Container, Head, Html, Preview } from "react-email";
import { EmailFooter } from "./footer";
import { EmailHeader } from "./header";

type EmailLayoutProps = {
  children: ReactNode;
  preview: string;
};

export function EmailLayout({ children, preview }: EmailLayoutProps) {
  return (
    <Html lang="en">
      <Head />
      <Preview>{preview}</Preview>
      <Body style={body}>
        <Container style={container}>
          <EmailHeader />
          {children}
          <EmailFooter />
        </Container>
      </Body>
    </Html>
  );
}

const body: CSSProperties = {
  margin: 0,
  backgroundColor: "#f5f5f5",
  color: "#171717",
  fontFamily: "Arial, Helvetica, sans-serif",
};

const container: CSSProperties = {
  width: "100%",
  maxWidth: "560px",
  margin: "40px auto",
  padding: "36px",
  backgroundColor: "#ffffff",
  border: "1px solid #e5e5e5",
  borderRadius: "12px",
};
