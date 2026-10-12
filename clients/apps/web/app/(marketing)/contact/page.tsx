import Link from "next/link";
import { PageIntro } from "@/components/marketing/page-intro";
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
      <PageIntro
        eyebrow="Contact"
        title="Let’s talk about your voice infrastructure."
      >
        <p>
          Find implementation guidance, report an issue, or start a conversation
          about Leamout.
        </p>
      </PageIntro>
      <div className="grid gap-6 pb-20 md:grid-cols-2">
        <section className="rounded-xl border border-border p-8">
          <h2 className="text-2xl font-semibold">Building with Leamout?</h2>
          <p className="my-5 leading-7 text-muted-foreground">
            Start with setup documentation and the project repository. Include
            your deployment environment and the behavior you are trying to
            achieve when asking a technical question.
          </p>
          <Link
            href="/docs"
            className="font-medium underline decoration-brand-orange underline-offset-8"
          >
            Read the documentation
          </Link>
        </section>
        <section className="rounded-xl border border-border p-8">
          <h2 className="text-2xl font-semibold">Questions and feedback</h2>
          <p className="my-5 leading-7 text-muted-foreground">
            Use the repository’s issue tracker for public questions and
            reproducible bugs. Do not include credentials, call recordings,
            personal data, or vulnerability details in public issues.
          </p>
          <a
            href="https://github.com/leamout/leamout/issues"
            className="font-medium underline decoration-brand-orange underline-offset-8"
          >
            Visit the issue tracker
          </a>
        </section>
      </div>
      <section className="mb-20 border-t border-border pt-8">
        <h2 className="text-xl font-semibold">
          Enterprise and private inquiries
        </h2>
        <p className="mt-4 max-w-2xl leading-7 text-muted-foreground">
          A dedicated private contact channel will be listed here once
          available.
        </p>
      </section>
    </div>
  );
}
