import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Subprocessors",
  description:
    "Review the third-party service providers that process personal data on behalf of Leamout.",
  path: "/subprocessors",
});

export default function page() {
  return <div>Subprocessors page</div>;
}
