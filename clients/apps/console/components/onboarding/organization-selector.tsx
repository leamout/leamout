"use client";

import { Button } from "@leamout/ui/components/button";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { errorMessage } from "@/lib/api";
import { useOrganization } from "./organization-provider";

export function OrganizationSelector() {
  const router = useRouter();
  const organization = useOrganization();
  const [error, setError] = useState("");

  function select(id: string) {
    setError("");
    if (!organization.selectOrganization(id)) {
      setError("Unable to save your organization selection. Please try again.");
      return;
    }
    router.replace("/");
  }

  const loading =
    !organization.ready ||
    organization.isPending ||
    organization.isFetching;

  return (
    <div className="space-y-6">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-semibold tracking-tight">
          Select an organization
        </h1>
        <p className="text-sm text-muted-foreground">
          Choose the workspace you want to manage.
        </p>
      </div>
      {organization.isError ? (
        <div className="space-y-3" role="alert">
          <p className="text-sm">{errorMessage(organization.error)}</p>
          <Button variant="outline" onClick={organization.refetch}>
            Try again
          </Button>
        </div>
      ) : loading ? (
        <p role="status" className="text-sm text-muted-foreground">
          Loading your organizations…
        </p>
      ) : organization.organizations.length > 0 ? (
        <ul className="space-y-3" aria-label="Organizations">
          {organization.organizations.map((item) => (
            <li key={item.id}>
              <Button
                className="h-auto w-full justify-between whitespace-normal break-words py-4 text-left"
                variant="outline"
                onClick={() => select(item.id)}
              >
                <span>{item.name}</span>
                {organization.selected?.id === item.id ? (
                  <span className="text-xs text-muted-foreground">Current</span>
                ) : null}
              </Button>
            </li>
          ))}
        </ul>
      ) : (
        <div className="rounded-xl border border-dashed p-6 text-center">
          <p className="font-medium">No organizations yet</p>
          <p className="mt-2 text-sm text-muted-foreground">
            Create an organization to configure your first Voice Agent.
          </p>
        </div>
      )}
      {error && <p role="alert" className="text-sm text-destructive">{error}</p>}
      <Button className="w-full" render={<Link href="/create-organization" />}>
        Create organization
      </Button>
    </div>
  );
}
