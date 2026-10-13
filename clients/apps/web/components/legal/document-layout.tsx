import type { ReactNode } from "react";
import { PageIntro } from "@/components/page-intro";

type DocumentLayoutProps = {
  title: string;
  description: string;
  children: ReactNode;
};

export function DocumentLayout({
  title,
  description,
  children,
}: DocumentLayoutProps) {
  return (
    <div className="mx-auto max-w-6xl px-5 sm:px-6 lg:px-8">
      <PageIntro eyebrow="Trust and transparency" title={title}>
        <p>{description}</p>
      </PageIntro>
      <div className="mb-10 inline-flex rounded-full border border-border px-4 py-2 font-mono text-xs uppercase tracking-wider">
        Pending publication
      </div>
      <div className="max-w-3xl pb-20">{children}</div>
    </div>
  );
}
