export function AboutIntro() {
  return (
    <div className="max-w-3xl py-16 sm:py-24">
      <p className="mb-5 font-mono text-xs uppercase tracking-[0.18em] text-brand-orange">
        About Leamout
      </p>
      <h1 className="text-balance font-heading text-4xl font-semibold leading-[1.08] tracking-tight sm:text-6xl">
        Infrastructure for the conversations ahead.
      </h1>
      <div className="mt-6 max-w-2xl text-pretty text-lg leading-8 text-muted-foreground">
        <p>
          We are building the runtime and control plane that developers use to build and run autonomous voice agents
          on Leamout Cloud.
        </p>
      </div>
    </div>
  );
}
