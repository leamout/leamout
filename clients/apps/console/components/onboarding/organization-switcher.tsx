"use client";

import { ArrowDown01Icon, Tick02Icon } from "@hugeicons/core-free-icons";
import { HugeiconsIcon } from "@hugeicons/react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@leamout/ui/components/dropdown-menu";
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@leamout/ui/components/sidebar";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useOrganization } from "./organization-provider";

export function OrganizationSwitcher() {
  const router = useRouter();
  const { isMobile, setOpenMobile } = useSidebar();
  const { organizations, selected, selectOrganization } = useOrganization();
  const [error, setError] = useState("");

  const activeOrganizations = organizations.filter(
    (organization) => organization.status === "active",
  );

  function switchOrganization(organizationId: string) {
    setError("");

    if (organizationId === selected?.id) {
      return;
    }

    if (!selectOrganization(organizationId)) {
      setError("Unable to switch organizations. Please try again.");
      return;
    }

    setOpenMobile(false);
    // Resource detail URLs belong to the previous organization.
    router.replace("/");
  }

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu onOpenChange={() => setError("")}>
          <DropdownMenuTrigger
            render={
              <SidebarMenuButton
                size="lg"
                tooltip="Switch organization"
                aria-label={
                  selected
                    ? "Switch organization, current: " + selected.name
                    : "Choose organization"
                }
                className="data-[popup-open]:bg-sidebar-accent data-[popup-open]:text-sidebar-accent-foreground"
              />
            }
          >
            <span className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-sidebar-primary text-sm font-semibold text-sidebar-primary-foreground">
              {selected?.name.charAt(0).toUpperCase() ?? "L"}
            </span>
            <span className="grid min-w-0 flex-1 text-left text-sm leading-tight group-data-[collapsible=icon]:hidden">
              <span className="truncate font-medium">
                {selected?.name ?? "Choose organization"}
              </span>
              <span className="truncate text-xs capitalize text-muted-foreground">
                {selected?.role ?? "Workspace"}
              </span>
            </span>
            <HugeiconsIcon
              icon={ArrowDown01Icon}
              size={16}
              strokeWidth={1.5}
              aria-hidden="true"
              className="ml-auto group-data-[collapsible=icon]:hidden"
            />
          </DropdownMenuTrigger>

          <DropdownMenuContent
            align="start"
            side={isMobile ? "bottom" : "right"}
            className="min-w-56 max-w-80"
          >
            <DropdownMenuLabel>Organizations</DropdownMenuLabel>
            {activeOrganizations.map((organization) => (
              <DropdownMenuItem
                key={organization.id}
                onClick={() => switchOrganization(organization.id)}
                className="gap-3"
              >
                <span className="flex size-8 shrink-0 items-center justify-center rounded-md bg-muted font-medium">
                  {organization.name.charAt(0).toUpperCase()}
                </span>
                <span className="grid min-w-0 flex-1 leading-tight">
                  <span className="truncate">{organization.name}</span>
                  <span className="truncate text-xs capitalize text-muted-foreground">
                    {organization.role}
                  </span>
                </span>
                {selected?.id === organization.id && (
                  <HugeiconsIcon
                    icon={Tick02Icon}
                    size={16}
                    strokeWidth={2}
                    aria-hidden="true"
                    className="ml-auto"
                  />
                )}
              </DropdownMenuItem>
            ))}
            {activeOrganizations.length === 0 && (
              <DropdownMenuItem disabled>
                No active organizations
              </DropdownMenuItem>
            )}
            <DropdownMenuSeparator />
            <DropdownMenuItem
              render={
                <Link
                  href="/create-organization"
                  onClick={() => setOpenMobile(false)}
                />
              }
            >
              <span className="text-lg" aria-hidden="true">
                +
              </span>
              Create organization
            </DropdownMenuItem>
            <DropdownMenuItem
              render={
                <Link
                  href="/organizations"
                  onClick={() => setOpenMobile(false)}
                />
              }
            >
              View all organizations
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
        {error && (
          <p
            role="alert"
            className="px-2 pt-2 text-xs text-destructive group-data-[collapsible=icon]:hidden"
          >
            {error}
          </p>
        )}
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
