import type { MetadataRoute } from "next";
import { posts } from "@/lib/blog/posts";

const paths = [
  "/",
  "/platform",
  "/integrations",
  "/pricing",
  "/about",
  "/security",
  "/contact",
  "/brand",
  "/blog",
];

export default function sitemap(): MetadataRoute.Sitemap {
  return [...paths, ...posts.map((post) => `/blog/${post.slug}`)].map((path) => ({
    url: new URL(path, "https://leamout.com").toString(),
  }));
}
