import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Cookie Policy",
  description:
    "Learn how Leamout uses cookies and similar technologies and how you can manage your preferences.",
  path: "/legal/cookie",
});

export default function page() {
  return <div>Cookie Policy page</div>;
}
