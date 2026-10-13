import type { ReactNode } from "react";

export function PageIntro({
  eyebrow,
  title,
  children,
}: {
  eyebrow: string;
  title: string;
  children: ReactNode;
}) {
  return (
    <div className="max-w-3xl py-16 sm:py-24">
      <p className="mb-5 font-mono text-xs uppercase tracking-[0.18em] text-brand-orange">
        {eyebrow}
      </p>
      <h1 className="text-balance font-heading text-4xl font-semibold leading-[1.08] tracking-tight sm:text-6xl">
        {title}
      </h1>
      <div className="mt-6 max-w-2xl text-pretty text-lg leading-8 text-muted-foreground">
        {children}
      </div>
    </div>
  );
}
