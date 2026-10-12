import { PageIntro } from "@/components/marketing/page-intro";
import { constructMetadata } from "@/lib/metadata";

export const metadata = constructMetadata({
  title: "Brand",
  description:
    "Find Leamout brand assets and guidelines for using our name, logo, and visual identity.",
  path: "/brand",
});

export default function BrandPage() {
  return (
    <div className="mx-auto max-w-6xl px-5 sm:px-6 lg:px-8">
      <PageIntro eyebrow="Brand" title="Leamout, clearly expressed.">
        <p>
          A simple visual identity for infrastructure that connects voice, AI,
          and applications.
        </p>
      </PageIntro>
      <section className="grid gap-4 sm:grid-cols-2">
        <div className="flex min-h-52 items-center justify-center rounded-xl border border-border bg-brand-white text-brand-charcoal">
          <span className="text-5xl font-semibold tracking-tight">
            Leamout<span className="text-brand-orange">.</span>
          </span>
        </div>
        <div className="flex min-h-52 items-center justify-center rounded-xl bg-brand-charcoal text-white">
          <span className="text-5xl font-semibold tracking-tight">
            Leamout<span className="text-brand-orange">.</span>
          </span>
        </div>
      </section>
      <section className="py-16">
        <h2 className="text-2xl font-semibold">Color palette</h2>
        <div className="mt-6 grid gap-4 sm:grid-cols-3">
          {[
            ["Orange", "#F43E01"],
            ["Charcoal", "#302B28"],
            ["White", "#FFFFFF"],
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
      <section className="grid gap-8 border-t border-border py-12 sm:grid-cols-2">
        <div>
          <h2 className="text-2xl font-semibold">Typography</h2>
          <p className="mt-4 text-2xl">IBM Plex Sans</p>
          <p className="mt-3 font-mono">IBM Plex Mono</p>
        </div>
        <div>
          <h2 className="text-2xl font-semibold">Writing the name</h2>
          <p className="mt-4 leading-8 text-muted-foreground">
            Write “Leamout” with an uppercase L. Keep the name readable and
            avoid implying a partnership, endorsement, or certification that has
            not been agreed.
          </p>
        </div>
      </section>
    </div>
  );
}
