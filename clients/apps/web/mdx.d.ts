declare module "*.mdx" {
  import type { ComponentType } from "react";

  export const metadata: {
    title: string;
    description: string;
    category: string;
    author: string;
    publishedAt: string;
    status: "draft" | "published";
  };

  const MDXContent: ComponentType;
  export default MDXContent;
}
