import type { ReactNode } from "react";

export function ArticleContent({ children }: { children: ReactNode }) {
  return <div className="mt-10 min-w-0 text-foreground">{children}</div>;
}
