import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Security",
  description:
    "Learn about security in Leamout, including access controls, credential handling, and deployment considerations for autonomous voice agents.",
  path: "/security",
});

export default function page() {
  return <div>Security page</div>;
}
