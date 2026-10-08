"use client";

import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarTrigger,
} from "@leamout/ui/components/sidebar";
import Link from "next/link";
import { usePathname } from "next/navigation";

const navigation = [
  { title: "Overview", href: "/" },
  { title: "Agents", href: "/agents" },
  { title: "Calls", href: "/calls" },
  { title: "Connections", href: "/connections" },
  { title: "Settings", href: "/settings" },
];

export function ConsoleShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const current =
    navigation.find((item) => item.href === pathname)?.title ?? "Console";

  return (
    <SidebarProvider>
      <Sidebar variant="inset" collapsible="offcanvas">
        <SidebarHeader className="gap-3 p-4">
          <Link href="/" className="text-lg font-semibold">
            Leamout
          </Link>
          <p className="text-xs text-muted-foreground">
            Organization selection coming soon
          </p>
        </SidebarHeader>
        <SidebarContent className="px-2">
          <SidebarMenu>
            {navigation.map((item) => (
              <SidebarMenuItem key={item.href}>
                <SidebarMenuButton
                  isActive={pathname === item.href}
                  render={<Link href={item.href} />}
                >
                  <span>{item.title}</span>
                </SidebarMenuButton>
              </SidebarMenuItem>
            ))}
          </SidebarMenu>
        </SidebarContent>
        <SidebarFooter className="p-4 text-xs text-muted-foreground">
          Console preview
        </SidebarFooter>
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
