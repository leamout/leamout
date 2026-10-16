"use client";

import { Button } from "@leamout/ui/components/button";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api, errorMessage } from "@/lib/api";
import { useSession } from "./require-session";

export function AccountMenu() {
  const session = useSession();
  const client = useQueryClient();
  const logout = useMutation({
    mutationFn: () =>
      api<{ logout_url: string }>("/auth/logout", {
        method: "POST",
        body: {},
      }),
    onSuccess: async ({ logout_url }) => {
      await client.cancelQueries();
      client.clear();
      window.location.assign(logout_url);
    },
  });

  return (
    <div className="space-y-2 px-2 py-2 group-data-[collapsible=icon]:hidden">
      <p className="truncate text-xs text-muted-foreground">
        {session.data?.email}
      </p>
      <Button
        variant="ghost"
        size="sm"
        disabled={logout.isPending}
        onClick={() => logout.mutate()}
      >
        {logout.isPending ? "Logging out…" : "Log out"}
      </Button>
      {logout.isError && (
        <p role="alert" className="text-xs text-destructive">
          {errorMessage(logout.error)}
        </p>
      )}
    </div>
  );
}
