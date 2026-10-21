import { breadcrumbsSchema } from "@/lib/seo/breadcrumbs";
import { JsonLd } from "@/components/seo/json-ld";
import { organizationSchema } from "@/lib/seo/organization";
import { webpageSchema } from "@/lib/seo/webpage";
import { websiteSchema } from "@/lib/seo/website";
import { AboutCallToAction } from "@/components/about/call-to-action";
import { AboutIntro } from "@/components/about/intro";
import { Mission } from "@/components/about/mission";
import { Principles } from "@/components/about/principles";
import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "About",
  description:
    "Why we are building Leamout: a managed, SIP-native runtime and control plane for autonomous voice agents.",
  path: "/about",
});

export default function AboutPage() {
  return (
      <div className="mx-auto max-w-[1200px] px-5 sm:px-6 lg:px-8">
      <AboutIntro />
      <Mission />
      <Principles />
      <AboutCallToAction />
    </div>
  </>
  );
}
