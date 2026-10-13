import type { ReactNode } from "react";

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
      <div className="max-w-3xl py-16 sm:py-24">
        <p className="mb-5 font-mono text-xs uppercase tracking-[0.18em] text-brand-orange">
          Trust and transparency
        </p>
        <h1 className="text-balance font-heading text-4xl font-semibold leading-[1.08] tracking-tight sm:text-6xl">
          {title}
        </h1>
        <div className="mt-6 max-w-2xl text-pretty text-lg leading-8 text-muted-foreground">
          <p>{description}</p>
        </div>
      </div>
      <div className="mb-10 inline-flex rounded-full border border-border px-4 py-2 font-mono text-xs uppercase tracking-wider">
        Pending publication
      </div>
      <div className="max-w-3xl pb-20">{children}</div>
    </div>
  );
}
