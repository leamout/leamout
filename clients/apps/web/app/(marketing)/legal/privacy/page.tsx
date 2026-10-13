import { PrivacyPolicy } from "@/components/legal/privacy-policy";
import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Privacy Policy",
  description:
    "Learn how Leamout collects, uses, and protects personal data, and understand your privacy rights.",
  noIndex: true,
  path: "/legal/privacy",
});

export default function PrivacyPolicyPage() {
  return <PrivacyPolicy />;
}
