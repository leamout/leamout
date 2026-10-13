export function BillingExplanation() {
  return (
    <section className="py-16">
      <h2 className="text-3xl font-semibold tracking-tight">
        Know what you are paying for.
      </h2>
      <div className="mt-8 grid gap-8 md:grid-cols-3">
        {[
          [
            "Leamout",
            "Your platform plan includes runtime capacity, agent and trunk limits, and retention. Pro and Scale also include advanced roles and configurable retention policies.",
          ],
          [
            "Your carrier",
            "You bring existing numbers and SIP connectivity. Your carrier bills you for phone numbers and telephony usage.",
          ],
          [
            "Your AI providers",
            "You supply your own provider credentials. Your AI providers bill you for speech and model usage.",
          ],
        ].map(([title, text]) => (
          <div key={title} className="border-t border-border pt-5">
            <h3 className="font-semibold">{title}</h3>
            <p className="mt-3 leading-7 text-muted-foreground">{text}</p>
          </div>
        ))}
      </div>
    </section>
  );
}
