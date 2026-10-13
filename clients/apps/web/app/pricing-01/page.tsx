import Pricing from "@/components/shadcn-space/blocks/pricing-01/pricing";
import { constructMetadata } from "@/lib/metadata";
import { consoleURL } from "@/lib/site";

export const metadata = constructMetadata({
  title: "Pricing block preview",
  description: "Preview the Leamout pricing block.",
  path: "/pricing",
  noIndex: true,
});

export default function Pricing01Page() {
  return (
    <main>
      <Pricing signupHref={`${consoleURL}/create-account`} />
    </main>
  );
}
