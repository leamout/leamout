"use client";

import { Button } from "@leamout/ui/components/button";
import { useRouter } from "next/navigation";
import { type ReactNode, useEffect } from "react";
import { errorMessage } from "@/lib/api";
import { useOrganization } from "./organization-provider";

export function RequireOrganization({ children }: { children: ReactNode }) {
  const router = useRouter();
  const organization = useOrganization();
  const loading =
    !organization.ready ||
    organization.isPending ||
    organization.isFetching;
  const redirect = !loading && !organization.isError && !organization.selected;

  useEffect(() => {
    if (redirect) router.replace("/organizations");
  }, [redirect, router]);

  if (organization.isError) {
    return (
      <div className="m-auto space-y-4 p-6" role="alert">
        <p>{errorMessage(organization.error)}</p>
        <Button onClick={organization.refetch}>Try again</Button>
      </div>
    );
  }

  if (loading || redirect) {
    return (
      <div className="m-auto p-6 text-sm text-muted-foreground" role="status">
        {redirect ? "Choose an organization to continue…" : "Loading organization…"}
      </div>
    );
  }

  return children;
}
