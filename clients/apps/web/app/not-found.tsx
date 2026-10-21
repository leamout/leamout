import { Button } from "@leamout/ui/components/button";
import Link from "next/link";
import { MarketingFooter } from "@/components/marketing/footer";
import { MarketingHeader } from "@/components/marketing/header";
import { consoleURL } from "@/lib/site";

export default function NotFound() {
  return (
    <div className="dark flex min-h-screen flex-col bg-background text-foreground">
      <MarketingHeader consoleURL={consoleURL} />

      <main
        id="main-content"
        tabIndex={-1}
        className="flex flex-1 items-center outline-none"
      >
        <section
          aria-labelledby="not-found-title"
          className="mx-auto w-full max-w-[1200px] px-5 py-20 sm:px-6 sm:py-24 lg:px-8 lg:py-28"
        >
          <div className="grid gap-14 lg:grid-cols-[1.05fr_0.95fr] lg:items-center lg:gap-16">
            <div>
              <p className="mb-7 flex items-center gap-3 font-mono text-xs font-medium uppercase tracking-[0.16em]">
                <span
                  aria-hidden="true"
                  className="size-2 rounded-full bg-brand-orange"
                />
                404 / Page not found
              </p>
              <h1
                id="not-found-title"
                className="max-w-2xl text-balance font-heading text-5xl font-semibold leading-[1.06] tracking-tight sm:text-6xl"
              >
                This route
                <br />
                <span className="text-muted-foreground">doesn't exist.</span>
              </h1>
              <p className="mt-8 max-w-xl text-lg leading-8 text-muted-foreground">
                The address may be incorrect, or the page may have moved. Head
                back to Leamout or explore the documentation to find what you
                need.
              </p>
              <div className="mt-9 flex flex-wrap gap-3">
                <Button
                  size="lg"
                  nativeButton={false}
                  render={<Link href="/" />}
                >
                  Back to home
                </Button>
                <Button
                  size="lg"
                  variant="outline"
                  nativeButton={false}
                  render={<Link href="/docs" />}
                >
                  Read documentation
                </Button>
              </div>
            </div>

            <div
              aria-hidden="true"
              className="flex min-h-72 flex-col justify-between overflow-hidden rounded-xl border border-border bg-brand-surface p-7 sm:min-h-96 sm:p-9"
            >
              <div className="flex items-center justify-between font-mono text-xs uppercase tracking-[0.14em] text-muted-foreground">
                <span>Request path</span>
                <span className="size-2 rounded-full bg-brand-orange" />
              </div>
              <p className="my-12 text-center font-heading text-[clamp(7rem,19vw,14rem)] font-semibold leading-none tracking-[-0.1em] text-brand-offwhite">
                4<span className="text-brand-orange">0</span>4
              </p>
              <div className="flex items-center justify-between gap-4 border-t border-border pt-5 font-mono text-xs uppercase tracking-[0.12em] text-muted-foreground">
                <span>No matching route</span>
                <span>Leamout</span>
              </div>
            </div>
          </div>
        </section>
      </main>

      <MarketingFooter />
    </div>
  );
}
