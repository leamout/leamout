"use client";

import { ArrowDown01Icon, Settings01Icon } from "@hugeicons/core-free-icons";
import { HugeiconsIcon } from "@hugeicons/react";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@leamout/ui/components/collapsible";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@leamout/ui/components/dropdown-menu";
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
  useSidebar,
} from "@leamout/ui/components/sidebar";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState } from "react";

const pages = [
  "Organization",
  "Members",
  "Profile",
  "Security",
  "API keys",
  "Storage",
].map((title) => ({
  title,
  href: `/settings/${title === "API keys" ? "api-keys" : title.toLowerCase()}`,
}));

export function SettingsSidebar() {
  const pathname = usePathname();
  return <SettingsGroup key={pathname} pathname={pathname} />;
}

function SettingsGroup({ pathname }: { pathname: string }) {
  const { state, isMobile, setOpenMobile } = useSidebar();
  const active = pathname === "/settings" || pathname.startsWith("/settings/");
  const [open, setOpen] = useState(active);
  const icon = (
    <HugeiconsIcon
      icon={Settings01Icon}
      size={20}
      strokeWidth={1.5}
      color="currentColor"
      aria-hidden="true"
    />
  );
  if (state === "collapsed" && !isMobile) {
    return (
      <SidebarMenu>
        <SidebarMenuItem>
          <DropdownMenu>
            <DropdownMenuTrigger
              render={
                <SidebarMenuButton
                  tooltip="Settings"
                  isActive={active}
                  aria-label="Settings"
                />
              }
            >
              {icon}
              <span>Settings</span>
            </DropdownMenuTrigger>
            <DropdownMenuContent side="right" align="end">
              {pages.map((page) => (
                <DropdownMenuItem
                  key={page.href}
                  render={
                    <Link
                      href={page.href}
                      aria-current={pathname === page.href ? "page" : undefined}
                    />
                  }
                  className={
                    pathname === page.href ? "bg-accent font-medium" : undefined
                  }
                >
                  {page.title}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
        </SidebarMenuItem>
      </SidebarMenu>
    );
  }
  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <Collapsible open={open} onOpenChange={setOpen}>
          <CollapsibleTrigger
            render={
              <SidebarMenuButton isActive={active} aria-label="Settings" />
            }
          >
            {icon}
            <span>Settings</span>
            <HugeiconsIcon
              icon={ArrowDown01Icon}
              size={16}
              aria-hidden="true"
              className={`ml-auto transition-transform ${open ? "rotate-180" : ""}`}
            />
          </CollapsibleTrigger>
          <CollapsibleContent>
            <SidebarMenuSub>
              {pages.map((page) => (
                <SidebarMenuSubItem key={page.href}>
                  <SidebarMenuSubButton
                    isActive={pathname === page.href}
                    render={
                      <Link
                        href={page.href}
                        aria-current={
                          pathname === page.href ? "page" : undefined
                        }
                      />
                    }
                    onClick={() => setOpenMobile(false)}
                  >
                    {page.title}
                  </SidebarMenuSubButton>
                </SidebarMenuSubItem>
              ))}
            </SidebarMenuSub>
          </CollapsibleContent>
        </Collapsible>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
