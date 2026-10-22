import { readdir } from "node:fs/promises";
import path from "node:path";
import { ArticleCard } from "@/components/blog/article-card";
import { constructMetadata } from "@/lib/metadata";
import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";

type Props = {
  searchParams: Promise<{ page?: string | string[] }>;
};

const PAGE_SIZE = 12;

function pageNumber(value: string | string[] | undefined): number {
  if (value === undefined) {
    return 1;
  }

  if (typeof value !== "string" || !/^[1-9]\d*$/.test(value)) {
    notFound();
  }

  const page = Number(value);

  if (!Number.isSafeInteger(page)) {
    notFound();
  }

  return page;
}

export async function generateMetadata({ searchParams }: Props): Promise<Metadata> {
  const page = pageNumber((await searchParams).page);

  return constructMetadata({
    title: page === 1 ? "Blog" : `Blog — Page ${page}`,
    description:
      "Engineering notes on SIP-native voice agents, real-time media, and the infrastructure behind Leamout.",
    path: page === 1 ? "/blog" : `/blog?page=${page}`,
  });
}

export default async function BlogPage({ searchParams }: Props) {
  const page = pageNumber((await searchParams).page);
  const directory = path.join(process.cwd(), "content");
  const files = await readdir(directory);

  const slugs = files
    .filter((file) => /^[a-z0-9]+(?:-[a-z0-9]+)*\.mdx$/.test(file))
    .map((file) => file.slice(0, -4));

  const posts = (
    await Promise.all(
      slugs.map(async (slug) => {
        const { metadata } = await import(`@/content/${slug}.mdx`);

        if (metadata.status !== "published") {
          return null;
        }

        return { slug, ...metadata };
      }),
    )
  )
    .filter((post): post is NonNullable<typeof post> => post !== null)
    .sort(
      (a, b) =>
        b.publishedAt.localeCompare(a.publishedAt) ||
        a.slug.localeCompare(b.slug),
    );

  const pageCount = Math.max(1, Math.ceil(posts.length / PAGE_SIZE));

  if (page > pageCount) {
    notFound();
  }

  const visiblePosts = posts.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE);

  return (
    <div className="mx-auto max-w-[1200px] px-5 py-20 sm:px-6 lg:px-8">
      <p className="font-mono text-xs uppercase tracking-widest text-brand-orange">
        Leamout / Writing
      </p>
      <h1 className="mt-5 font-heading text-5xl font-semibold tracking-tight">
        Ideas behind the infrastructure.
      </h1>
      <p className="mt-6 max-w-2xl text-lg text-muted-foreground">
        Engineering notes, architecture decisions, and lessons from building
        infrastructure for autonomous voice agents.
      </p>
      <div className="mt-16">
        {visiblePosts.map((post) => (
          <ArticleCard key={post.slug} post={post} />
        ))}
      </div>
      <nav
        aria-label="Blog pages"
        className="mt-10 flex items-center justify-between border-t border-border pt-8"
      >
        {page > 1 ? (
          <Link
            href={page === 2 ? "/blog" : `/blog?page=${page - 1}`}
            rel="prev"
            className="text-sm hover:text-brand-orange"
          >
            Previous
          </Link>
        ) : (
          <span />
        )}
        <span className="text-sm text-muted-foreground">
          Page {page} of {pageCount}
        </span>
        {page < pageCount ? (
          <Link
            href={`/blog?page=${page + 1}`}
            rel="next"
            className="text-sm hover:text-brand-orange"
          >
            Next
          </Link>
        ) : (
          <span />
        )}
      </nav>
    </div>
  );
}
