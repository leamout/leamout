import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Terms of Service",
  description:
    "Read the terms governing access to and use of Leamout services.",
  path: "/legal/terms",
});

export default function page() {
  return <div>Terms of Service page</div>;
}
