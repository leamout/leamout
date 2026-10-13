import { Implementation } from "@/components/security/implementation";
import { SecurityIntro } from "@/components/security/intro";
import { SecurityControls } from "@/components/security/security-controls";
import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Security",
  description:
    "Learn about security in Leamout, including access controls, credential handling, and deployment considerations for autonomous voice agents.",
  path: "/security",
});

export default function SecurityPage() {
  return (
    <div className="mx-auto max-w-6xl px-5 sm:px-6 lg:px-8">
      <SecurityIntro />
      <SecurityControls />
      <Implementation />
    </div>
  );
}
