"use client";

import { Button } from "@leamout/ui/components/button";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@leamout/ui/components/collapsible";
import {
  NavigationMenu,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
  navigationMenuTriggerStyle,
} from "@leamout/ui/components/navigation-menu";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useRef, useState } from "react";

const links = [
  { label: "Platform", href: "/platform" },
  { label: "Integrations", href: "/integrations" },
  { label: "Pricing", href: "/pricing" },
  { label: "Docs", href: "/docs" },
];

export function MarketingHeader({ consoleURL }: { consoleURL: string }) {
  const pathname = usePathname();
  const [open, setOpen] = useState(false);
  const trigger = useRef<HTMLButtonElement>(null);

  return (
    <header className="sticky top-0 z-50 border-b border-border bg-background/95 backdrop-blur-xl">
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-50 focus:rounded-lg focus:bg-background focus:p-3 focus:ring-2 focus:ring-ring"
      >
        Skip to content
      </a>
      <Collapsible open={open} onOpenChange={setOpen}>
        <div className="mx-auto flex h-[76px] max-w-[1200px] items-center justify-between gap-6 px-5 sm:px-6 lg:px-8">
          <Link
            href="/"
            aria-label="Leamout home"
            className="rounded-sm font-heading text-xl font-semibold tracking-tight outline-none focus-visible:ring-2 focus-visible:ring-ring"
            onClick={() => setOpen(false)}
          >
            Leamout
            <span aria-hidden="true" className="text-brand-orange">
              .
            </span>
          </Link>
          <NavigationMenu
            aria-label="Main navigation"
            className="hidden lg:flex"
          >
            <NavigationMenuList>
              {links.map((link) => (
                <NavigationMenuItem key={link.href}>
                  <NavigationMenuLink
                    render={<Link href={link.href} />}
                    active={pathname === link.href}
                    className={navigationMenuTriggerStyle()}
                  >
                    {link.label}
                  </NavigationMenuLink>
                </NavigationMenuItem>
              ))}
            </NavigationMenuList>
          </NavigationMenu>
          <div className="hidden items-center gap-2 lg:flex">
            <Button
              variant="ghost"
              nativeButton={false}
              render={<a href={`${consoleURL}/log-in`} />}
            >
              Log in
            </Button>
            <Button
              nativeButton={false}
              render={<a href={`${consoleURL}/create-account`} />}
            >
              Get started
            </Button>
          </div>
          <CollapsibleTrigger
            ref={trigger}
            render={<Button variant="outline" className="lg:hidden" />}
          >
            {open ? "Close menu" : "Menu"}
          </CollapsibleTrigger>
        </div>
        <CollapsibleContent
          className="border-t border-border lg:hidden"
          onKeyDown={(event) => {
            if (event.key === "Escape") {
              setOpen(false);
              trigger.current?.focus();
            }
          }}
        >
          <nav
            aria-label="Mobile navigation"
            className="mx-auto flex max-w-[1200px] flex-col gap-2 px-5 py-5 sm:px-6 lg:px-8"
          >
            {links.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                aria-current={pathname === link.href ? "page" : undefined}
                onClick={() => setOpen(false)}
                className="rounded-lg px-3 py-3 text-sm font-medium outline-none hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring aria-[current=page]:bg-muted"
              >
                {link.label}
              </Link>
            ))}
            <div className="mt-3 grid grid-cols-2 gap-3 border-t border-border pt-5">
              <Button
                variant="outline"
                nativeButton={false}
                render={<a href={`${consoleURL}/log-in`} />}
              >
                Log in
              </Button>
              <Button
                nativeButton={false}
                render={<a href={`${consoleURL}/create-account`} />}
              >
                Get started
              </Button>
            </div>
          </nav>
        </CollapsibleContent>
      </Collapsible>
    </header>
  );
}
