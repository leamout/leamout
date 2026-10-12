import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Contact",
  description:
    "Contact Leamout to discuss autonomous voice infrastructure, self-hosting, and enterprise requirements.",
  path: "/contact",
});

export default function page() {
  return <div>Contact page</div>;
}
