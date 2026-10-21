import { JsonLd } from "@/components/seo/json-ld";
import { organizationSchema } from "@/lib/seo/organization";
import { webpageSchema } from "@/lib/seo/webpage";
import { websiteSchema } from "@/lib/seo/website";
import { breadcrumbsSchema } from "@/lib/seo/breadcrumbs";
import { Implementation } from "@/components/security/implementation";
import { SecurityIntro } from "@/components/security/intro";
import { SecurityControls } from "@/components/security/security-controls";
import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Security",
  description:
    "Explore Leamout's approach to organization boundaries, provider credentials, call security, and responsible vulnerability disclosure.",
  path: "/security",
});

export default function SecurityPage() {
  return (
    <>
      <JsonLd nodes={[organizationSchema(), websiteSchema(), webpageSchema({ path: "/security", name: "Leamout Security", description: "Leamout's approach to organization boundaries, provider credentials, call security, and vulnerability disclosure." }), breadcrumbsSchema([{name:"Home",path:"/"},{name:"Security",path:"/security"}])]} />
      <div className="mx-auto max-w-[1200px] px-5 sm:px-6 lg:px-8">
      <SecurityIntro />
      <SecurityControls />
      <Implementation />
    </div>
    </>
  );
}
