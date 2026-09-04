import type { ReactNode } from "react";
import {
  Body,
  Container,
  Head,
  Html,
  Preview,
  Section,
  Tailwind,
} from "react-email";
import { leamoutTailwindConfig } from "../theme";
import { EmailFooter } from "./footer";
import { LeamoutFonts } from "./fonts";
import { EmailHeader } from "./header";

type EmailLayoutProps = {
  children: ReactNode;
  preview: string;
};

export function EmailLayout({ children, preview }: EmailLayoutProps) {
  return (
    <Tailwind config={leamoutTailwindConfig}>
      <Html lang="en">
        <Head>
          <LeamoutFonts />
        </Head>
        <Body className="m-0 bg-canvas p-0 font-sans text-fg">
          <Preview>{preview}</Preview>
          <Container className="mx-auto max-w-[640px] px-4 pt-12 pb-6 sm:pt-16">
            <Section className="rounded-[8px] shadow-matte-card">
              <Section className="rounded-[8px] border border-stroke bg-bg">
                <Section className="px-6 pt-12 sm:px-10 sm:pt-16">
                  <EmailHeader />
                </Section>
                {children}
                <EmailFooter />
              </Section>
            </Section>
          </Container>
        </Body>
      </Html>
    </Tailwind>
  );
}
