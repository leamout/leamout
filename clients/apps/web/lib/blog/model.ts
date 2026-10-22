export type BlogPost = {
  slug: string;
  title: string;
  description: string;
  category: "Engineering" | "Architecture" | "Building Leamout" | "Company";
  publishedAt: string;
  author: string;
};
