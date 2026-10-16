export type Organization = {
  id: string;
  name: string;
  status: string;
  role: "owner" | "admin" | "member";
  created_at?: string;
  updated_at?: string;
};

export function organizationStorageKey(userId: string): string {
  return `leamout:organization:${userId}`;
}

export function readOrganizationSelection(
  storage: Pick<Storage, "getItem">,
  userId: string,
): string | null {
  try {
    return storage.getItem(organizationStorageKey(userId));
  } catch {
    return null;
  }
}

export function writeOrganizationSelection(
  storage: Pick<Storage, "setItem">,
  userId: string,
  organizationId: string,
): boolean {
  try {
    storage.setItem(organizationStorageKey(userId), organizationId);
    return true;
  } catch {
    return false;
  }
}

export function findActiveOrganization(
  organizations: Organization[] | undefined,
  organizationId: string | null,
): Organization | null {
  return (
    organizations?.find(
      (organization) =>
        organization.status === "active" && organization.id === organizationId,
    ) ?? null
  );
}
