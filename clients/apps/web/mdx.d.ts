declare module "*.mdx" {
  import type { BlogMetadata } from "@/lib/blog";
  import type { ComponentType } from "react";

  export const metadata: BlogMetadata;

  const MDXContent: ComponentType;
  export default MDXContent;
}
