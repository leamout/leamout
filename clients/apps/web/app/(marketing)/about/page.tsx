import { AboutCallToAction } from "@/components/about/call-to-action";
import { AboutIntro } from "@/components/about/intro";
import { Mission } from "@/components/about/mission";
import { Principles } from "@/components/about/principles";
import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "About",
  description:
    "Learn about Leamout and our mission to help developers run autonomous voice agents through Leamout Cloud using their own carriers and AI providers.",
  path: "/about",
});

export default function AboutPage() {
  return (
    <div className="mx-auto max-w-6xl px-5 sm:px-6 lg:px-8">
      <AboutIntro />
      <Mission />
      <Principles />
      <AboutCallToAction />
    </div>
  );
}
