const questions = [
  {
    question: "Are carrier minutes or phone numbers included?",
    answer:
      "No. Leamout is the managed voice agent runtime, not your phone carrier. Bring existing numbers and SIP connectivity; your carrier bills those services separately.",
  },
  {
    question: "How does AI provider billing work?",
    answer:
      "With your own AI provider credentials, speech and model usage is billed by those providers. If your organization uses Leamout-provided AI, any applicable usage charges are separate from your subscription.",
  },
  {
    question: "What happens when I use all my included runtime minutes?",
    answer:
      "New voice agent sessions are rejected when your monthly allowance is exhausted. Calls already in progress can finish. There is no automatic runtime top-up.",
  },
  {
    question: "Do I have to host the voice agent infrastructure myself?",
    answer:
      "No. Leamout is one managed service. We operate the runtime and control plane; you connect your SIP carrier, AI providers, and application tools.",
  },
  {
    question: "Can I start small and increase capacity later?",
    answer:
      "The Free, Developer, Pro, and Scale subscriptions offer progressively higher runtime, concurrency, agent, and SIP trunk limits. Contact us if you need a different capacity profile.",
  },
];

export function PricingFAQ() {
  return (
    <section className="border-t border-border py-20" aria-labelledby="pricing-faq-title">
      <div className="grid gap-10 lg:grid-cols-[0.75fr_1.25fr] lg:gap-20">
        <div>
          <p className="font-mono text-xs uppercase tracking-[0.16em] text-brand-orange">
            Frequently asked
          </p>
          <h2
            id="pricing-faq-title"
            className="mt-5 text-balance font-heading text-3xl font-semibold tracking-tight sm:text-4xl"
          >
            Questions about pricing.
          </h2>
          <p className="mt-4 max-w-sm leading-7 text-muted-foreground">
            The essentials of running and paying for Leamout.
          </p>
        </div>
        <div className="border-t border-border">
          {questions.map(({ question, answer }) => (
            <details key={question} className="group border-b border-border py-5">
              <summary className="flex cursor-pointer list-none items-center justify-between gap-6 text-left font-medium leading-7 outline-none focus-visible:ring-2 focus-visible:ring-brand-orange [&::-webkit-details-marker]:hidden">
                <span>{question}</span>
                <span
                  aria-hidden="true"
                  className="text-xl font-normal text-brand-orange transition-transform group-open:rotate-45 motion-reduce:transition-none"
                >
                  +
                </span>
              </summary>
              <p className="mt-4 max-w-xl pr-8 text-sm leading-7 text-muted-foreground">
                {answer}
              </p>
            </details>
          ))}
        </div>
      </div>
    </section>
  );
}
