import { ContactOptions } from "@/components/contact/contact-options";
import { ContactIntro } from "@/components/contact/intro";
import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Contact",
  description:
    "Find Leamout documentation, product feedback channels, and a private way to report security concerns.",
  path: "/contact",
});

export default function ContactPage() {
  return (
    <div className="mx-auto max-w-[1200px] px-5 sm:px-6 lg:px-8">
      <ContactIntro />
      <ContactOptions />
    </div>
  );
}
