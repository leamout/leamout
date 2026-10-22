import "server-only";

import { access, readdir } from "node:fs/promises";
import path from "node:path";
import { cache } from "react";

const contentDirectory = path.join(process.cwd(), "content");
const slugPattern = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;

export type BlogMetadata = {
  title: string;
  description: string;
  category: string;
  author: string;
  publishedAt: string;
  status: "draft" | "published";
};

export type BlogPost = BlogMetadata & {
  slug: string;
};

function validateMetadata(slug: string, metadata: BlogMetadata): void {
  if (
    !metadata ||
    [
      metadata.title,
      metadata.description,
      metadata.category,
      metadata.author,
    ].some((value) => typeof value !== "string" || value.trim().length === 0) ||
    !/^\d{4}-\d{2}-\d{2}$/.test(metadata.publishedAt) ||
    !["draft", "published"].includes(metadata.status)
  ) {
    throw new Error(`Invalid metadata in content/${slug}.mdx`);
  }

  const date = new Date(`${metadata.publishedAt}T00:00:00.000Z`);
  if (
    !Number.isFinite(date.getTime()) ||
    date.toISOString().slice(0, 10) !== metadata.publishedAt
  ) {
    throw new Error(`Invalid publication date in content/${slug}.mdx`);
  }
}

const loadModule = cache(async (slug: string) => {
  return import(`@/content/${slug}.mdx`);
});

export const getPost = cache(async (slug: string) => {
  if (!slugPattern.test(slug)) {
    return null;
  }

  try {
    await access(path.join(contentDirectory, `${slug}.mdx`));
  } catch (error) {
    if ((error as NodeJS.ErrnoException).code === "ENOENT") {
      return null;
    }

    throw error;
  }

  const { default: Component, metadata } = await loadModule(slug);
  validateMetadata(slug, metadata);

  if (metadata.status !== "published") {
    return null;
  }

  return { slug, metadata, Component };
});

export const getPosts = cache(async (): Promise<BlogPost[]> => {
  const files = await readdir(contentDirectory);
  const slugs = files
    .filter((file) => file.endsWith(".mdx"))
    .map((file) => file.slice(0, -4))
    .filter((slug) => slugPattern.test(slug))
    .sort();

  const posts: BlogPost[] = [];

  for (let start = 0; start < slugs.length; start += 32) {
    const batch = await Promise.all(
      slugs.slice(start, start + 32).map((slug) => getPost(slug)),
    );

    for (const post of batch) {
      if (post) {
        posts.push({ slug: post.slug, ...post.metadata });
      }
    }
  }

  return posts.sort(
    (a, b) =>
      b.publishedAt.localeCompare(a.publishedAt) ||
      a.slug.localeCompare(b.slug),
  );
});

export async function getPostSlugs(): Promise<{ slug: string }[]> {
  const posts = await getPosts();
  return posts.map(({ slug }) => ({ slug }));
}
