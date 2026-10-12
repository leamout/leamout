import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Pricing",
  description:
    "Explore Leamout platform subscriptions for running autonomous voice agents with your own carriers and AI providers.",
  path: "/pricing",
});

export default function page() {
  return <div>Pricing page</div>;
}
