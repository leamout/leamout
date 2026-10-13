import { ColorPalette } from "@/components/brand/color-palette";
import { BrandIntro } from "@/components/brand/intro";
import { Typography } from "@/components/brand/typography";
import { Wordmark } from "@/components/brand/wordmark";
import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Brand",
  description:
    "Find Leamout brand assets and guidelines for using our name, logo, and visual identity.",
  path: "/brand",
});

export default function BrandPage() {
  return (
    <div className="mx-auto max-w-6xl px-5 sm:px-6 lg:px-8">
      <BrandIntro />
      <Wordmark />
      <ColorPalette />
      <Typography />
    </div>
  );
}
