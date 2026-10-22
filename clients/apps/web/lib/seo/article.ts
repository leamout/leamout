import { siteUrl, type SchemaNode } from "./schema";
import type { BlogPost } from "@/lib/blog/model";

export function articleSchema(post: BlogPost): SchemaNode {
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
