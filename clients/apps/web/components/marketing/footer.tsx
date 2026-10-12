import Link from "next/link";

const groups = [
  {
    title: "Resources",
    links: [
      { label: "Pricing", href: "/pricing" },
      { label: "Documentation", href: "/docs" },
      { label: "Security", href: "/security" },
    ],
  },
  {
    title: "Company",
    links: [
      { label: "About", href: "/about" },
      { label: "Brand", href: "/brand" },
      { label: "Contact", href: "/contact" },
    ],
  },
  {
    title: "Legal",
    links: [
      { label: "Privacy", href: "/legal/privacy" },
      { label: "Terms", href: "/legal/terms" },
      { label: "Cookies", href: "/legal/cookie" },
      { label: "Subprocessors", href: "/subprocessors" },
    ],
  },
];

export function MarketingFooter() {
  return (
    <footer className="border-t border-border bg-muted/20">
      <div className="mx-auto max-w-6xl px-5 py-12 sm:px-6 lg:px-8">
        <div className="grid gap-10 lg:grid-cols-[1.2fr_2fr] lg:gap-20">
          <div>
            <Link
              href="/"
              aria-label="Leamout home"
              className="rounded-sm font-heading text-xl font-semibold tracking-tight outline-none focus-visible:ring-2 focus-visible:ring-ring"
            >
              Leamout
              <span aria-hidden="true" className="text-brand-orange">
                .
              </span>
            </Link>
            <p className="mt-4 max-w-sm text-sm leading-6 text-muted-foreground">
              Leamout is the carrier-grade runtime and control plane for
              autonomous voice agents.
            </p>
          </div>
          <div className="grid grid-cols-2 gap-8 sm:grid-cols-3">
            {groups.map((group) => (
              <nav key={group.title} aria-label={group.title}>
                <h2 className="text-sm font-semibold">{group.title}</h2>
                <ul className="mt-4 space-y-3">
                  {group.links.map((link) => (
                    <li key={link.href}>
                      <Link
                        href={link.href}
                        className="rounded-sm text-sm text-muted-foreground outline-none transition-colors hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
                      >
                        {link.label}
                      </Link>
                    </li>
                  ))}
                </ul>
              </nav>
            ))}
          </div>
        </div>
        <p className="mt-12 border-t border-border pt-6 text-xs text-muted-foreground">
          © {new Date().getFullYear()} Leamout. All rights reserved.
        </p>
      </div>
    </footer>
  );
}
