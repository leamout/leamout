import type { BlogPost } from "@/lib/blog";
import type { TechArticle } from "schema-dts";

import { siteUrl } from "./schema";

export function articleSchema(post: BlogPost): TechArticle {
  return {
    "@type": "TechArticle",
    "@id": `${siteUrl(`/blog/${post.slug}`)}#article`,
    headline: post.title,
    description: post.description,
    datePublished: post.publishedAt,
    mainEntityOfPage: siteUrl(`/blog/${post.slug}`),
    author: { "@id": "https://vidzro.com/#person" },
    publisher: { "@id": "https://leamout.com/#organization" },
  };
}
