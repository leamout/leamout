"use client";

import {
  AiBrain01Icon,
  Call02Icon,
  DashboardSquare01Icon,
  LinkSquare02Icon,
  RoboticIcon,
  TelephoneIcon,
  ToolsIcon,
  VoiceIcon,
  WebhookIcon,
} from "@hugeicons/core-free-icons";
import { HugeiconsIcon } from "@hugeicons/react";
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
import { SettingsSidebar } from "@/components/settings/settings-sidebar";

const navigation = [
  { title: "Overview", href: "/", icon: DashboardSquare01Icon },
  { title: "Agents", href: "/agents", icon: RoboticIcon },
  { title: "Calls", href: "/calls", icon: Call02Icon },
  { title: "Recordings", href: "/recordings", icon: VoiceIcon },
  { title: "Phone numbers", href: "/phone-numbers", icon: TelephoneIcon },
  { title: "SIP trunks", href: "/sip-trunks", icon: LinkSquare02Icon },
  { title: "AI providers", href: "/ai-providers", icon: AiBrain01Icon },
  { title: "Tools", href: "/tools", icon: ToolsIcon },
  { title: "Webhooks", href: "/webhooks", icon: WebhookIcon },
];

function Navigation() {
  const pathname = usePathname();
  const { setOpenMobile } = useSidebar();

  return (
    <SidebarMenu>
      {navigation.map((item) => {
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
              <HugeiconsIcon
                icon={item.icon}
                size={20}
                color="currentColor"
                strokeWidth={1.5}
                aria-hidden="true"
              />
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
    )?.title ?? (pathname.startsWith("/settings") ? "Settings" : "Console");

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
            <SettingsSidebar />
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
