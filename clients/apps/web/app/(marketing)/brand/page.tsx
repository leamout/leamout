import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Brand",
  description:
    "Find Leamout brand assets and guidelines for using our name, logo, and visual identity.",
  path: "/brand",
});

export default function page() {
  return <div>Brand page</div>;
}
