import type { BlogPost } from "@/lib/blog";
import Link from "next/link";

export function ArticleCard({ post }: { post: BlogPost }) {
  return (
    <article className="border-t border-border py-8">
      <p className="font-mono text-xs uppercase tracking-wider text-brand-orange">
        {post.category} · <time dateTime={post.publishedAt}>{post.publishedAt}</time>
      </p>
      <h2 className="mt-3 font-heading text-2xl font-semibold tracking-tight">
        <Link href={`/blog/${post.slug}`} className="hover:text-brand-orange">
          {post.title}
        </Link>
      </h2>
      <p className="mt-3 max-w-2xl text-muted-foreground">
        {post.description}
      </p>
      <p className="mt-4 text-sm text-muted-foreground">{post.author}</p>
    </article>
  );
}
