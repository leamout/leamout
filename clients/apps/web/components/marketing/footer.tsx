import Link from "next/link";

const groups = [
  {
    title: "Resources",
    links: [
      { label: "Platform", href: "/platform" },
      { label: "Integrations", href: "/integrations" },
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
    <footer className="border-t border-border bg-[#26211F]">
      <div className="mx-auto max-w-[1200px] px-5 pt-16 pb-8 sm:px-6 lg:px-8">
        <div className="grid gap-12 lg:grid-cols-[1fr_1.5fr] lg:gap-20">
          <div>
            <Link
              href="/"
              aria-label="Leamout home"
              className="rounded-sm font-heading text-2xl font-semibold tracking-tight outline-none focus-visible:ring-2 focus-visible:ring-ring"
            >
              Leamout
              <span aria-hidden="true" className="text-brand-orange">
                .
              </span>
            </Link>
            <p className="mt-5 max-w-sm leading-7 text-muted-foreground">
              The carrier-grade runtime and control plane for autonomous voice
              agents.
            </p>
            <p className="mt-7 font-mono text-[11px] uppercase tracking-[0.15em] text-[#FF8E6D]">
              SIP-native. Provider-independent.
            </p>
          </div>
          <div className="grid grid-cols-2 gap-10 sm:grid-cols-3">
            {groups.map((group) => (
              <nav key={group.title} aria-label={group.title}>
                <h2 className="font-mono text-xs font-medium uppercase tracking-wider text-foreground">
                  {group.title}
                </h2>
                <ul className="mt-5 space-y-3">
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
        <div className="mt-16 flex flex-col gap-3 border-t border-border pt-6 text-xs text-muted-foreground sm:flex-row sm:justify-between">
          <p>© {new Date().getFullYear()} Leamout. All rights reserved.</p>
          <p>One managed runtime. Your carrier and AI providers.</p>
        </div>
      </div>
    </footer>
  );
}
