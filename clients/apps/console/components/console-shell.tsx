"use client";

import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarRail,
  SidebarTrigger,
  useSidebar,
} from "@leamout/ui/components/sidebar";
import Link from "next/link";
import { usePathname } from "next/navigation";

const navigation = [
  { title: "Overview", href: "/" },
  { title: "Agents", href: "/agents" },
  { title: "Calls", href: "/calls" },
  { title: "Phone numbers", href: "/phone-numbers" },
  { title: "SIP trunks", href: "/sip-trunks" },
  { title: "AI providers", href: "/ai-providers" },
  { title: "Tools", href: "/tools" },
  { title: "Webhooks", href: "/webhooks" },
  { title: "Settings", href: "/settings" },
];

function Navigation({ settings = false }: { settings?: boolean }) {
  const pathname = usePathname();
  const { setOpenMobile } = useSidebar();
  const paths = [
    "M3 3h7v7H3z M14 3h7v7h-7z M3 14h7v7H3z M14 14h7v7h-7z",
    "M5 7h14v13H5z M12 3v4 M8 12h1 M15 12h1 M9 16h6",
    "M5 3h4l2 5-3 2a14 14 0 0 0 6 6l2-3 5 2v4C10 21 3 14 3 5z",
    "M9 15l6-6 M8 13l-2 2a3 3 0 0 0 4 4l3-3 M16 11l2-2a3 3 0 0 0-4-4l-3 3",
    "M4 4h16v16H4z M8 8h8 M8 12h8 M8 16h4",
    "M12 3v18 M3 12h18 M5 5l14 14 M5 19L19 5",
    "M4 7h16 M4 17h16 M8 4v6 M16 14v6",
    "M5 5h14v14H5z M9 9h6 M9 13h6",
    "M4 7h16 M4 17h16 M8 4v6 M16 14v6",
  ];

  return (
    <SidebarMenu>
      {navigation.map((item, index) => {
        if ((item.href === "/settings") !== settings) return null;
        const active =
          item.href === "/"
            ? pathname === "/"
            : pathname === item.href || pathname.startsWith(`${item.href}/`);
        return (
          <SidebarMenuItem key={item.href}>
            <SidebarMenuButton
              isActive={active}
              tooltip={item.title}
              render={
                <Link
                  href={item.href}
                  aria-label={item.title}
                  aria-current={active ? "page" : undefined}
                />
              }
              onClick={() => setOpenMobile(false)}
            >
              <svg
                aria-hidden="true"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="1.6"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d={paths[index]} />
              </svg>
              <span>{item.title}</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        );
      })}
    </SidebarMenu>
  );
}

export function ConsoleShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const current =
    navigation.find((item) =>
      item.href === "/"
        ? pathname === "/"
        : pathname === item.href || pathname.startsWith(`${item.href}/`),
    )?.title ?? "Console";

  return (
    <SidebarProvider>
      <Sidebar variant="inset" collapsible="icon">
        <SidebarHeader className="gap-3 p-2">
          <Link
            href="/"
            className="flex items-center gap-2 text-lg font-semibold"
            aria-label="Leamout overview"
          >
            <span className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
              L
            </span>
            <span className="group-data-[collapsible=icon]:hidden">
              Leamout
            </span>
          </Link>
          <p className="px-2 text-xs text-muted-foreground group-data-[collapsible=icon]:hidden">
            Organization selection coming soon
          </p>
        </SidebarHeader>
        <SidebarContent>
          <SidebarGroup>
            <SidebarGroupContent>
              <nav aria-label="Main navigation">
                <Navigation />
              </nav>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>
        <SidebarFooter>
          <nav aria-label="Settings navigation">
            <Navigation settings />
          </nav>
          <p className="px-2 py-2 text-xs text-muted-foreground group-data-[collapsible=icon]:hidden">
            Console preview · Account integration coming soon
          </p>
        </SidebarFooter>
        <SidebarRail />
      </Sidebar>
      <SidebarInset>
        <header className="flex h-16 items-center gap-3 border-b px-6">
          <SidebarTrigger />
          <span className="text-sm font-medium">{current}</span>
        </header>
        <main className="mx-auto w-full max-w-6xl p-6 md:p-10">{children}</main>
      </SidebarInset>
    </SidebarProvider>
  );
}
