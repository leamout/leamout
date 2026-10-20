export function ColorPalette() {
  return (
    <section className="py-16">
      <h2 className="text-2xl font-heading font-semibold">Color palette</h2>
      <div className="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {[
          ["Graphite", "#101316"],
          ["Surface", "#1C2226"],
          ["Off-white", "#F5F6F4"],
          ["Leamout Orange", "#F43E01"],
        ].map(([name, color]) => (
          <div
            key={name}
            className="overflow-hidden rounded-xl border border-border"
          >
            <div className="h-28" style={{ backgroundColor: color }} />
            <div className="flex justify-between p-4 text-sm">
              <span>{name}</span>
              <code>{color}</code>
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}
