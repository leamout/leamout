import type { BlogPost } from "./model";

export const posts: BlogPost[] = [
  {
    slug: "why-sip-native-voice-agents",
    title: "Why SIP-native infrastructure matters for voice agents",
    description:
      "An engineering perspective on connecting existing telephony systems to real-time AI without surrendering call control.",
    category: "Architecture",
    publishedAt: "2026-10-22",
    author: "Coffey Vidzro",
  },
];

export function getPost(slug: string): BlogPost | undefined {
  return posts.find((post) => post.slug === slug);
}
