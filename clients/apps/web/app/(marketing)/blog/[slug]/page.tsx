import { JsonLd } from "@/components/seo/json-ld";
import { getPost, posts } from "@/lib/blog/posts";
import { articleSchema } from "@/lib/seo/article";
import { founderSchema } from "@/lib/seo/person";
import { organizationSchema } from "@/lib/seo/organization";
import { constructMetadata } from "@/lib/metadata";
import type { Metadata } from "next";
import { notFound } from "next/navigation";

type Props = { params: Promise<{ slug: string }> };

export const dynamicParams = false;

export function generateStaticParams() {
  return posts.map(({ slug }) => ({ slug }));
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params;
  const post = getPost(slug);
  if (!post) return {};
  return constructMetadata({
    title: post.title,
    description: post.description,
    path: `/blog/${slug}`,
  });
}

export default async function BlogArticlePage({ params }: Props) {
  const { slug } = await params;
  const post = getPost(slug);
  if (!post) notFound();

  // Explicit imports allow Next.js to bundle MDX at build time.
  if (slug !== "why-sip-native-voice-agents") notFound();
  const { default: Article } = await import(
    "@/content/blog/why-sip-native-voice-agents.mdx"
  );

  return (
    <article className="mx-auto max-w-3xl px-5 py-20 sm:px-6">
      <JsonLd nodes={[organizationSchema(), founderSchema(), articleSchema(post)]} />
      <p className="font-mono text-xs uppercase tracking-wider text-brand-orange">
        {post.category} · {post.publishedAt}
      </p>
      <p className="mt-3 text-sm text-muted-foreground">By {post.author}</p>
      <div className="prose prose-invert mt-10 max-w-none">
        <Article />
      </div>
    </article>
  );
}
