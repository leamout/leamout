"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

const pages = [
  "Organization",
  "Members",
  "Profile",
  "Security",
  "API keys",
  "Storage",
];

export function SettingsNavigation() {
  const pathname = usePathname();
  return (
    <nav
      aria-label="Settings sections"
      className="flex flex-wrap gap-2 border-b pb-4"
    >
      {pages.map((title) => {
        const href = `/settings/${title === "API keys" ? "api-keys" : title.toLowerCase()}`;
        return (
          <Link
            key={href}
            href={href}
            aria-current={pathname === href ? "page" : undefined}
            className={`rounded-md px-3 py-2 text-sm ${pathname === href ? "bg-muted font-medium" : "text-muted-foreground hover:bg-muted"}`}
          >
            {title}
          </Link>
        );
      })}
    </nav>
  );
}
