import type { BlogPost } from "@/lib/blog";

export function ArticleHeader({ post }: { post: BlogPost }) {
  const date = new Date(`${post.publishedAt}T00:00:00.000Z`).toLocaleDateString(
    "en-US",
    { year: "numeric", month: "long", day: "numeric", timeZone: "UTC" },
  );

  return (
    <header className="border-b border-border pb-10">
      <p className="font-mono text-xs uppercase tracking-wider text-brand-orange">
        {post.category}
      </p>
      <h1 className="mt-5 font-heading text-4xl font-semibold tracking-tight sm:text-5xl">
        {post.title}
      </h1>
      <p className="mt-5 text-lg leading-8 text-muted-foreground">
        {post.description}
      </p>
      <div className="mt-6 flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-muted-foreground">
        <a href="https://vidzro.com" className="underline underline-offset-4">
          {post.author}
        </a>
        <span aria-hidden="true">·</span>
        <time dateTime={post.publishedAt}>{date}</time>
      </div>
    </header>
  );
}
