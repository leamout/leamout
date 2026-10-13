import { TermsOfService } from "@/components/legal/terms-of-service";
import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Terms of Service",
  description:
    "Read the terms governing access to and use of Leamout services.",
  noIndex: true,
  path: "/legal/terms",
});

export default function TermsOfServicePage() {
  return <TermsOfService />;
}
