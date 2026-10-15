import { ContactOptions } from "@/components/contact/contact-options";
import { ContactIntro } from "@/components/contact/intro";
import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Contact",
  description:
    "Contact Leamout about using Leamout Cloud for autonomous voice agents.",
  path: "/contact",
});

export default function ContactPage() {
  return (
    <div className="mx-auto max-w-6xl px-5 sm:px-6 lg:px-8">
      <ContactIntro />
      <ContactOptions />
    </div>
  );
}
