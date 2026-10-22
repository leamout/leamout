import { ArticleContent } from "@/components/blog/article-content";
import { ArticleHeader } from "@/components/blog/article-header";
import { JsonLd } from "@/components/seo/json-ld";
import { getPost, getPosts } from "@/lib/blog";
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
  const posts = await getPosts();
  return posts.map(({ slug }) => ({ slug }));
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params;
  const post = await getPost(slug);

  if (!post) {
    notFound();
  }

  const metadata = constructMetadata({
    title: post.title,
    description: post.description,
    path: `/blog/${post.slug}`,
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
  const post = await getPost(slug);

  if (!post) {
    notFound();
  }

  // The slug has been validated against the published MDX files.
  // Next.js bundles matching local MDX modules at build time.
  const { default: Article } = await import(`@/content/blog/${post.slug}.mdx`);

  return (
    <article className="mx-auto max-w-3xl px-5 py-20 sm:px-6">
      <JsonLd
        nodes={[organizationSchema(), founderSchema(), articleSchema(post)]}
      />
      <ArticleHeader post={post} />
      <ArticleContent>
        <Article />
      </ArticleContent>
    </article>
  );
}
