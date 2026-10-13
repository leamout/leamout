import { Subprocessors } from "@/components/legal/subprocessors";
import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Subprocessors",
  description:
    "Review the third-party service providers that process personal data on behalf of Leamout.",
  noIndex: true,
  path: "/subprocessors",
});

export default function SubprocessorsPage() {
  return <Subprocessors />;
}
