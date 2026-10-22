import { JsonLd } from "@/components/seo/json-ld";
import { getPost, posts } from "@/lib/blog/posts";
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

export function generateStaticParams() {
  return posts.map(({ slug }) => ({ slug }));
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params;
  const post = getPost(slug);

  if (!post) {
    return {};
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
  const post = getPost(slug);

  if (!post) {
    notFound();
  }

  // Keep the MDX import statically analyzable for the Next.js compiler.
  if (slug !== "why-sip-native-voice-agents") {
    notFound();
  }

  const { default: Article } = await import(
    "@/content/blog/why-sip-native-voice-agents.mdx"
  );

  return (
    <article className="mx-auto max-w-3xl px-5 py-20 sm:px-6">
      <JsonLd
        nodes={[organizationSchema(), founderSchema(), articleSchema(post)]}
      />
      <p className="font-mono text-xs uppercase tracking-wider text-brand-orange">
        {post.category} · <time dateTime={post.publishedAt}>{post.publishedAt}</time>
      </p>
      <p className="mt-3 text-sm text-muted-foreground">
        By <a className="underline underline-offset-4" href="https://vidzro.com">{post.author}</a>
      </p>
      <div className="mt-10">
        <Article />
      </div>
    </article>
  );
}
