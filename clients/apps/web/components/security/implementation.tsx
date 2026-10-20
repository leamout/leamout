import { Button } from "@leamout/ui/components/button";
import Link from "next/link";

export function Implementation() {
  return (
    <section className="py-20">
      <div className="grid gap-10 rounded-xl border border-border bg-brand-surface p-8 sm:p-10 lg:grid-cols-[1.2fr_0.8fr] lg:items-end lg:gap-16 lg:p-12">
        <div>
          <p className="font-mono text-xs uppercase tracking-[0.16em] text-brand-orange">
            Responsible disclosure
          </p>
          <h2 className="mt-5 max-w-xl font-heading text-3xl font-semibold tracking-tight sm:text-4xl">
            Help us keep the platform secure.
          </h2>
          <p className="mt-5 max-w-xl leading-7 text-muted-foreground">
            Please report suspected vulnerabilities privately to
            security@leamout.com. Include reproduction steps and redacted
            diagnostics, without production credentials or customer media.
          </p>
        </div>
        <div className="flex flex-wrap gap-3 lg:justify-end">
          <Button
            nativeButton={false}
            render={<a href="mailto:security@leamout.com" />}
          >
            Report privately
          </Button>
          <Button
            variant="outline"
            nativeButton={false}
            render={<Link href="/docs" />}
          >
            Review documentation
          </Button>
        </div>
      </div>
    </section>
  );
}
