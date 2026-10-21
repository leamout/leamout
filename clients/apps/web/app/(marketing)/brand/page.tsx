import { ColorPalette } from "@/components/brand/color-palette";
import { BrandIntro } from "@/components/brand/intro";
import { Typography } from "@/components/brand/typography";
import { Wordmark } from "@/components/brand/wordmark";
import { JsonLd } from "@/components/seo/json-ld";
import { constructMetadata } from "@/lib/metadata";
import { breadcrumbsSchema } from "@/lib/seo/breadcrumbs";
import { organizationSchema } from "@/lib/seo/organization";
import { webpageSchema } from "@/lib/seo/webpage";
import { websiteSchema } from "@/lib/seo/website";

export const metadata = constructMetadata({
  title: "Brand",
  description:
    "Explore the Leamout identity: Graphite, Off-white, Leamout Orange, the wordmark, and our typography.",
  path: "/brand",
});

export default function BrandPage() {
  return (
    <>
      <JsonLd
        nodes={[
          organizationSchema(),
          websiteSchema(),
          webpageSchema({
            path: "/brand",
            name: "Leamout Brand",
            description:
              "Explore Leamout’s brand identity, colors, wordmark, and typography.",
          }),
          breadcrumbsSchema([
            { name: "Home", path: "/" },
            { name: "Brand", path: "/brand" },
          ]),
        ]}
      />
      <div className="mx-auto max-w-[1200px] px-5 sm:px-6 lg:px-8">
        <BrandIntro />
        <Wordmark />
        <ColorPalette />
        <Typography />
      </div>
    </>
  );
}
