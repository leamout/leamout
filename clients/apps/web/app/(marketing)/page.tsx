import { HomeCallToAction } from "@/components/home/call-to-action";
import { Capabilities } from "@/components/home/capabilities";
import { HomeHero } from "@/components/home/hero";
import { HowItWorks } from "@/components/home/how-it-works";
import { Ownership } from "@/components/home/ownership";

export default function Home() {
  return (
    <div>
      <HomeHero />
      <HowItWorks />
      <Capabilities />
      <Ownership />
      <HomeCallToAction />
    </div>
  );
}
