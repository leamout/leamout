import { ArticleCard } from "@/components/blog/article-card";
import { posts } from "@/lib/blog/posts";
import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Blog",
  description:
    "Engineering notes on SIP-native voice agents, real-time media, and the infrastructure behind Leamout.",
  path: "/blog",
});

export default function BlogPage() {
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
        {posts.map((post) => (
          <ArticleCard key={post.slug} post={post} />
        ))}
      </div>
    </div>
  );
}
