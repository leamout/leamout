declare module "*.mdx" {
  import type { ComponentType } from "react";
  import type { BlogMetadata } from "@/lib/blog";

  export const metadata: BlogMetadata;

  const MDXContent: ComponentType;
  export default MDXContent;
}
