import { ArticleContent } from "@/components/blog/article-content";
import { ArticleHeader } from "@/components/blog/article-header";
import { JsonLd } from "@/components/seo/json-ld";
import { getPost, getPostSlugs } from "@/lib/blog";
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

export const generateStaticParams = getPostSlugs;

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params;
  const post = await getPost(slug);

  if (!post) {
    notFound();
  }

  const metadata = constructMetadata({
    title: post.metadata.title,
    description: post.metadata.description,
    path: `/blog/${slug}`,
  });

  return {
    ...metadata,
    authors: [{ name: post.metadata.author, url: "https://vidzro.com" }],
    openGraph: {
      ...metadata.openGraph,
      type: "article",
      publishedTime: post.metadata.publishedAt,
      authors: [post.metadata.author],
    },
  };
}

export default async function BlogArticlePage({ params }: Props) {
  const { slug } = await params;
  const post = await getPost(slug);

  if (!post) {
    notFound();
  }

  const Post = post.Component;

  return (
    <article className="mx-auto max-w-3xl px-5 py-20 sm:px-6">
      <JsonLd
        nodes={[
          organizationSchema(),
          founderSchema(),
          articleSchema(slug, post.metadata),
        ]}
      />
      <ArticleHeader metadata={post.metadata} />
      <ArticleContent>
        <Post />
      </ArticleContent>
    </article>
  );
}
