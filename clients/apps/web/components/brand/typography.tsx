export function Typography() {
  return (
    <section className="grid gap-8 border-t border-border py-12 sm:grid-cols-2">
      <div>
        <h2 className="text-2xl font-heading font-semibold">Typography</h2>
        <p className="mt-4 font-heading text-2xl font-semibold">
          Space Grotesk
        </p>
        <p className="mt-3">IBM Plex Sans</p>
        <p className="mt-3 font-mono text-sm">IBM Plex Mono</p>
      </div>
      <div>
        <h2 className="text-2xl font-heading font-semibold">Writing the name</h2>
        <p className="mt-4 leading-8 text-muted-foreground">
          Write “Leamout” with an uppercase L. Keep the name readable and avoid
          implying a partnership, endorsement, or certification that has not
          been agreed.
        </p>
      </div>
    </section>
  );
}
