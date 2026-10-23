"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createContext,
  type ReactNode,
  useContext,
  useEffect,
  useState,
} from "react";
import { useSession } from "@/components/auth/require-session";
import { api } from "@/lib/api";
import {
  findActiveOrganization,
  type Organization,
  organizationStorageKey,
  readOrganizationSelection,
  writeOrganizationSelection,
} from "@/lib/organizations";

type OrganizationContextValue = {
  organizations: Organization[];
  selected: Organization | null;
  ready: boolean;
  isPending: boolean;
  isFetching: boolean;
  isError: boolean;
  error: Error | null;
  refetch: () => void;
  selectOrganization: (organizationId: string) => boolean;
  addOrganization: (organization: Organization) => boolean;
};

const OrganizationContext = createContext<OrganizationContextValue | null>(
  null,
);

export function OrganizationProvider({ children }: { children: ReactNode }) {
  const session = useSession();
  const userId = session.data?.id ?? "";
  const queryClient = useQueryClient();
  const [stored, setStored] = useState<{
    userId: string;
    organizationId: string | null;
  } | null>(null);

  const queryKey = ["organizations", userId];
  const organizations = useQuery({
    queryKey,
    queryFn: ({ signal }) => api<Organization[]>("/organizations", { signal }),
    enabled: userId !== "",
    staleTime: 0,
  });

  useEffect(() => {
    if (!userId) return;

    const sync = () => {
      setStored({
        userId,
        organizationId: readOrganizationSelection(window.localStorage, userId),
      });
    };

    sync();
    const onStorage = (event: StorageEvent) => {
      if (event.key !== organizationStorageKey(userId) && event.key !== null) {
        return;
      }

      // The selection may have changed in another tab. Discard all tenant
      // data before rendering the newly selected organization.
      void queryClient.cancelQueries({
        predicate: (query) =>
          query.queryKey[0] !== "session" &&
          query.queryKey[0] !== "organizations",
      });
      queryClient.removeQueries({
        predicate: (query) =>
          query.queryKey[0] !== "session" &&
          query.queryKey[0] !== "organizations",
      });
      sync();
      void queryClient.invalidateQueries({
        queryKey: ["organizations", userId],
      });

      // Detail routes may refer to resources from the previous organization.
      window.location.replace("/");
    };

    window.addEventListener("storage", onStorage);
    return () => window.removeEventListener("storage", onStorage);
  }, [userId, queryClient]);

  const ready = userId !== "" && stored?.userId === userId;
  const selected = findActiveOrganization(
    organizations.data,
    ready ? (stored?.organizationId ?? null) : null,
  );

  function selectOrganization(organizationId: string): boolean {
    const available = queryClient.getQueryData<Organization[]>(queryKey);
    if (!userId || !findActiveOrganization(available, organizationId)) {
      return false;
    }

    if (
      !writeOrganizationSelection(window.localStorage, userId, organizationId)
    ) {
      return false;
    }

    void queryClient.cancelQueries({
      predicate: (query) =>
        query.queryKey[0] !== "session" &&
        query.queryKey[0] !== "organizations",
    });
    setStored({ userId, organizationId });
    // Never reuse organization-scoped cached data when switching tenants.
    queryClient.removeQueries({
      predicate: (query) =>
        query.queryKey[0] !== "session" &&
        query.queryKey[0] !== "organizations",
    });
    return true;
  }

  function addOrganization(organization: Organization): boolean {
    if (!userId || organization.status !== "active") return false;

    queryClient.setQueryData<Organization[]>(queryKey, (current = []) => [
      organization,
      ...current.filter((item) => item.id !== organization.id),
    ]);
    return selectOrganization(organization.id);
  }

  return (
    <OrganizationContext.Provider
      value={{
        organizations: organizations.data ?? [],
        selected,
        ready,
        isPending: organizations.isPending,
        isFetching: organizations.isFetching,
        isError: organizations.isError,
        error: organizations.error,
        refetch: () => {
          void organizations.refetch();
        },
        selectOrganization,
        addOrganization,
      }}
    >
      {children}
    </OrganizationContext.Provider>
  );
}

export function useOrganization() {
  const context = useContext(OrganizationContext);
  if (!context) {
    throw new Error("useOrganization requires OrganizationProvider");
  }
  return context;
}
