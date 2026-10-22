import "server-only";

import { readdir, readFile } from "node:fs/promises";
import path from "node:path";
import { cache } from "react";
import matter from "gray-matter";

const directory = path.join(process.cwd(), "content", "blog");
const slugPattern = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;

export const BLOG_PAGE_SIZE = 12;

export type BlogPost = {
  slug: string;
  title: string;
  description: string;
  category: string;
  author: string;
  publishedAt: string;
  status: "draft" | "published";
};

function requiredString(data: Record<string, unknown>, name: string, slug: string): string {
  const value = data[name];

  if (typeof value !== "string" || value.trim().length === 0) {
    throw new Error(`Invalid blog metadata "${name}" in ${slug}.mdx`);
  }

  return value.trim();
}

function parsePost(slug: string, source: string): BlogPost {
  const { data } = matter(source);
  const metadata = data as Record<string, unknown>;
  const publishedAt = requiredString(metadata, "publishedAt", slug);
  const status = requiredString(metadata, "status", slug);

  if (!/^\d{4}-\d{2}-\d{2}$/.test(publishedAt)) {
    throw new Error(`Invalid publication date in ${slug}.mdx (expected YYYY-MM-DD)`);
  }

  const date = new Date(`${publishedAt}T00:00:00.000Z`);

  if (!Number.isFinite(date.getTime()) || date.toISOString().slice(0, 10) !== publishedAt) {
    throw new Error(`Invalid calendar date in ${slug}.mdx`);
  }

  if (status !== "draft" && status !== "published") {
    throw new Error(`Invalid blog status in ${slug}.mdx`);
  }

  return {
    slug,
    title: requiredString(metadata, "title", slug),
    description: requiredString(metadata, "description", slug),
    category: requiredString(metadata, "category", slug),
    author: requiredString(metadata, "author", slug),
    publishedAt,
    status,
  };
}

async function loadPost(slug: string): Promise<BlogPost | undefined> {
  if (!slugPattern.test(slug)) {
    return undefined;
  }

  try {
    const source = await readFile(path.join(directory, `${slug}.mdx`), "utf8");
    return parsePost(slug, source);
  } catch (error) {
    if ((error as NodeJS.ErrnoException).code === "ENOENT") {
      return undefined;
    }

    throw error;
  }
}

export const getPost = cache(async (slug: string): Promise<BlogPost | undefined> => {
  const post = await loadPost(slug);
  return post?.status === "published" ? post : undefined;
});

export const getPosts = cache(async (): Promise<BlogPost[]> => {
  const files = (await readdir(directory))
    .filter((file) => file.endsWith(".mdx"))
    .sort();

  const results: BlogPost[] = [];

  // Bound filesystem concurrency for large article collections.
  for (let start = 0; start < files.length; start += 32) {
    const batch = files.slice(start, start + 32);

    const posts = await Promise.all(
      batch.map((file) => loadPost(file.slice(0, -4))),
    );

    for (const post of posts) {
      if (post?.status === "published") {
        results.push(post);
      }
    }
  }

  return results.sort(
    (a, b) =>
      b.publishedAt.localeCompare(a.publishedAt) ||
      a.slug.localeCompare(b.slug),
  );
});

export function blogPage(posts: BlogPost[], page: number) {
  const pageCount = Math.max(1, Math.ceil(posts.length / BLOG_PAGE_SIZE));

  return {
    posts: posts.slice((page - 1) * BLOG_PAGE_SIZE, page * BLOG_PAGE_SIZE),
    pageCount,
  };
}
