import { readdir } from "node:fs/promises";
import path from "node:path";
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
  const directory = path.join(process.cwd(), "content");
  const files = await readdir(directory);

  const slugs = files
    .filter((file) => /^[a-z0-9]+(?:-[a-z0-9]+)*\.mdx$/.test(file))
    .map((file) => file.slice(0, -4));

  const articlePaths = await Promise.all(
    slugs.map(async (slug) => {
      const { metadata } = await import(`@/content/${slug}.mdx`);
      return metadata.status === "published" ? `/blog/${slug}` : null;
    }),
  );

  return [...paths, ...articlePaths.filter((path): path is string => path !== null)].map(
    (path) => ({
      url: new URL(path, "https://leamout.com").toString(),
    }),
  );
}
