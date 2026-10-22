import { readdir } from "node:fs/promises";
import path from "node:path";
import { ArticleContent } from "@/components/blog/article-content";
import { ArticleHeader } from "@/components/blog/article-header";
import { JsonLd } from "@/components/seo/json-ld";
import { constructMetadata } from "@/lib/metadata";
import { articleSchema } from "@/lib/seo/article";
import { organizationSchema } from "@/lib/seo/organization";
import { founderSchema } from "@/lib/seo/person";
import type { Metadata } from "next";
import { notFound } from "next/navigation";

type Props = {
  params: Promise<{ slug: string }>;
};

export const dynamicParams = false;

export async function generateStaticParams() {
  const directory = path.join(process.cwd(), "content");
  const files = await readdir(directory);

  const slugs = files
    .filter((file) => /^[a-z0-9]+(?:-[a-z0-9]+)*\.mdx$/.test(file))
    .map((file) => file.slice(0, -4));

  const posts = await Promise.all(
    slugs.map(async (slug) => {
      const { metadata } = await import(`@/content/${slug}.mdx`);
      return metadata.status === "published" ? { slug } : null;
    }),
  );

  return posts.filter((post): post is { slug: string } => post !== null);
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params;

  if (!/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(slug)) {
    notFound();
  }

  const { metadata: post } = await import(`@/content/${slug}.mdx`);

  if (post.status !== "published") {
    notFound();
  }

  const metadata = constructMetadata({
    title: post.title,
    description: post.description,
    path: `/blog/${slug}`,
  });

  return {
    ...metadata,
    authors: [{ name: post.author, url: "https://vidzro.com" }],
    openGraph: {
      ...metadata.openGraph,
      type: "article",
      publishedTime: post.publishedAt,
      authors: [post.author],
    },
  };
}

export default async function BlogArticlePage({ params }: Props) {
  const { slug } = await params;

  if (!/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(slug)) {
    notFound();
  }

  const { default: Post, metadata } = await import(`@/content/${slug}.mdx`);

  if (metadata.status !== "published") {
    notFound();
  }

  return (
    <article className="mx-auto max-w-3xl px-5 py-20 sm:px-6">
      <JsonLd
        nodes={[
          organizationSchema(),
          founderSchema(),
          articleSchema(slug, metadata),
        ]}
      />
      <ArticleHeader metadata={metadata} />
      <ArticleContent>
        <Post />
      </ArticleContent>
    </article>
  );
}
