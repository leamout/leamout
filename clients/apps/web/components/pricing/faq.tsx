export function PricingFAQ() {
  return (
    <section className="border-t border-border py-12">
      <h2 className="text-2xl font-semibold">Common questions</h2>
      {[
        [
          "Are carrier and AI costs included?",
          "No. Leamout does not resell carrier minutes or AI usage. Those charges remain with the providers you choose.",
        ],
        [
          "What happens when my runtime allowance is exhausted?",
          "New voice-agent sessions are rejected once the monthly allowance is exhausted. Active calls can finish; there is no automatic usage top-up.",
        ],
        [
          "Can I use my existing numbers?",
          "Yes. Connect them through your SIP carrier, PBX, or SBC. Number purchasing is not part of Leamout.",
        ],
        [
          "Can I self-host?",
          "Leamout supports self-hosted deployment. Read the documentation for setup requirements and contact us to discuss enterprise deployment support.",
        ],
      ].map(([question, answer]) => (
        <details key={question} className="border-b border-border py-5">
          <summary className="cursor-pointer font-medium focus-visible:outline-brand-orange">
            {question}
          </summary>
          <p className="mt-4 max-w-3xl leading-7 text-muted-foreground">
            {answer}
          </p>
        </details>
      ))}
    </section>
  );
}
