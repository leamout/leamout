import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Privacy Policy",
  description:
    "Learn how Leamout collects, uses, and protects personal data, and understand your privacy rights.",
  path: "/legal/privacy",
});

export default function page() {
  return <div>Privacy Policy page</div>;
}
