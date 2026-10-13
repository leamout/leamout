import { CookiePolicy } from "@/components/legal/cookie-policy";
import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Cookie Policy",
  description:
    "Learn how Leamout uses cookies and similar technologies and how you can manage your preferences.",
  noIndex: true,
  path: "/legal/cookie",
});

export default function CookiePolicyPage() {
  return <CookiePolicy />;
}
