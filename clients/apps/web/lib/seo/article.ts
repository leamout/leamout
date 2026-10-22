import type { BlogMetadata } from "@/lib/blog";
import type { TechArticle } from "schema-dts";

import { siteUrl } from "./schema";

export function articleSchema(
  slug: string,
  metadata: BlogMetadata,
): TechArticle {
  return {
    "@type": "TechArticle",
    "@id": `${siteUrl(`/blog/${slug}`)}#article`,
    headline: metadata.title,
    description: metadata.description,
    datePublished: metadata.publishedAt,
    mainEntityOfPage: siteUrl(`/blog/${slug}`),
    author: { "@id": "https://vidzro.com/#person" },
    publisher: { "@id": "https://leamout.com/#organization" },
  };
}
