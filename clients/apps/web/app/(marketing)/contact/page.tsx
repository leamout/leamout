import { ContactOptions } from "@/components/contact/contact-options";
import { EnterpriseInquiries } from "@/components/contact/enterprise-inquiries";
import { ContactIntro } from "@/components/contact/intro";
import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Contact",
  description:
    "Contact Leamout to discuss autonomous voice infrastructure, self-hosting, and enterprise requirements.",
  path: "/contact",
});

export default function ContactPage() {
  return (
    <div className="mx-auto max-w-6xl px-5 sm:px-6 lg:px-8">
      <ContactIntro />
      <ContactOptions />
      <EnterpriseInquiries />
    </div>
  );
}
