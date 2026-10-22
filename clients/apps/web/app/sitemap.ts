import { getPosts } from "@/lib/blog";
import type { MetadataRoute } from "next";

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

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const posts = await getPosts();
  const articlePaths = posts.map((post) => `/blog/${post.slug}`);

  return [...paths, ...articlePaths].map((path) => ({
    url: new URL(path, "https://leamout.com").toString(),
  }));
}
