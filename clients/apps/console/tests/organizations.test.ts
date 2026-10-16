import assert from "node:assert/strict";
import { test } from "node:test";
import {
  findActiveOrganization,
  organizationStorageKey,
  readOrganizationSelection,
  writeOrganizationSelection,
  type Organization,
} from "../lib/organizations";

const organizations: Organization[] = [
  { id: "org-a", name: "Alpha", status: "active", role: "owner" },
  { id: "org-b", name: "Beta", status: "active", role: "member" },
  { id: "org-disabled", name: "Disabled", status: "disabled", role: "owner" },
];

test("selection is scoped to the authenticated user", () => {
  const values = new Map<string, string>();
  const storage = {
    getItem: (key: string) => values.get(key) ?? null,
    setItem: (key: string, value: string) => {
      values.set(key, value);
    },
  };

  assert.equal(writeOrganizationSelection(storage, "user-a", "org-a"), true);
  assert.equal(readOrganizationSelection(storage, "user-a"), "org-a");
  assert.equal(readOrganizationSelection(storage, "user-b"), null);
  assert.equal(organizationStorageKey("user-a"), "leamout:organization:user-a");
});

test("only active organizations in the fetched membership list can be selected", () => {
  assert.equal(findActiveOrganization(organizations, "org-b")?.name, "Beta");
  assert.equal(findActiveOrganization(organizations, "org-disabled"), null);
  assert.equal(findActiveOrganization(organizations, "unrelated"), null);
  assert.equal(findActiveOrganization([], "org-a"), null);
  assert.equal(findActiveOrganization(undefined, "org-a"), null);
  assert.equal(findActiveOrganization(organizations, null), null);
});

test("unavailable browser storage cannot crash onboarding", () => {
  const storage = {
    getItem: (_key: string): string | null => {
      throw new Error("storage blocked");
    },
    setItem: (_key: string, _value: string) => {
      throw new Error("storage blocked");
    },
  };
  assert.equal(readOrganizationSelection(storage, "user-a"), null);
  assert.equal(writeOrganizationSelection(storage, "user-a", "org-a"), false);
});
