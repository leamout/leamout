import Link from "next/link";

export function ContactOptions() {
  return (
    <div className="grid gap-6 pb-20 md:grid-cols-2">
      <section className="rounded-xl border border-border p-8">
        <h2 className="text-2xl font-semibold">Building with Leamout?</h2>
        <p className="my-5 leading-7 text-muted-foreground">
          Start with setup documentation and the project repository. Include
          your deployment environment and the behavior you are trying to achieve
          when asking a technical question.
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
  );
}
