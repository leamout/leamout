import type { ReactNode } from "react";
import { Body, Container, Head, Html, Preview, Tailwind } from "react-email";
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
      <Tailwind>
        <Body className="m-0 bg-neutral-100 font-sans text-neutral-900">
          <Container className="mx-auto my-10 w-full max-w-[560px] rounded-xl border border-neutral-200 bg-white p-9">
            <EmailHeader />
            {children}
            <EmailFooter />
          </Container>
        </Body>
      </Tailwind>
    </Html>
  );
}
