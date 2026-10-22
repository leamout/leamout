const colors = [
  { name: "Graphite", hex: "#101316", use: "Foundation and deep backgrounds" },
  { name: "Surface", hex: "#1C2226", use: "Panels and layered content" },
  { name: "Off-white", hex: "#F5F6F4", use: "Text and light canvases" },
  {
    name: "Leamout Orange",
    hex: "#F43E01",
    use: "Actions and deliberate accents",
  },
];

export function ColorPalette() {
  return (
    <section
      className="border-b border-border py-20"
      aria-labelledby="brand-colors-title"
    >
      <div className="mb-9 flex flex-col justify-between gap-5 md:flex-row md:items-end">
        <div>
          <p className="font-mono text-xs uppercase tracking-[0.16em] text-brand-orange">
            02 / Color
          </p>
          <h2
            id="brand-colors-title"
            className="mt-5 font-heading text-3xl font-semibold tracking-tight sm:text-4xl"
          >
            Four colors. One system.
          </h2>
        </div>
        <p className="max-w-sm leading-7 text-muted-foreground">
          Graphite and Off-white carry the experience. Orange identifies
          decisions, actions, and details that matter.
        </p>
      </div>
      <div className="grid gap-px overflow-hidden rounded-xl border border-border bg-border sm:grid-cols-2 lg:grid-cols-4">
        {colors.map(({ name, hex, use }) => (
          <div key={name} className="bg-card">
            <div
              className="h-44 border-b border-border"
              style={{ backgroundColor: hex }}
              role="img"
              aria-label={`${name} color sample`}
            />
            <div className="space-y-2 p-6">
              <div className="flex flex-wrap items-baseline justify-between gap-2">
                <h3 className="font-heading text-lg font-medium">{name}</h3>
                <code className="font-mono text-xs text-muted-foreground">
                  {hex}
                </code>
              </div>
              <p className="text-sm leading-6 text-muted-foreground">{use}</p>
            </div>
          </div>
        ))}
      </div>
      <p className="mt-5 text-sm leading-7 text-muted-foreground">
        These colors are shared design tokens across Leamout products, not a
        separate palette for the marketing site.
      </p>
    </section>
  );
}
