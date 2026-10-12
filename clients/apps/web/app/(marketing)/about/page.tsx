import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "About",
  description:
    "Learn about Leamout and our mission to help developers and enterprises run autonomous voice agents with their own carriers and AI providers.",
  path: "/about",
});

export default function page() {
  return <div>About page</div>;
}
